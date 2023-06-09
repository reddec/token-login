package main

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/hashicorp/go-multierror"
	"github.com/jessevdk/go-flags"
	"github.com/jmoiron/sqlx"
	oidclogin "github.com/reddec/oidc-login"
	"golang.org/x/crypto/bcrypt"

	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/internal/dbo/pg"
	"github.com/reddec/token-login/internal/dbo/sqllite"
	"github.com/reddec/token-login/internal/validator"
	"github.com/reddec/token-login/web"
	"github.com/reddec/token-login/web/controllers/utils"
)

//nolint:gochecknoglobals
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

type Config struct {
	Admin Server `group:"Admin server configuration" namespace:"admin" env-namespace:"ADMIN"`
	Auth  Server `group:"Auth server configuration" namespace:"auth" env-namespace:"AUTH"`
	Login string `long:"login" env:"LOGIN" description:"Login method for admin UI" default:"basic" choice:"basic" choice:"oidc"`
	OIDC  OIDC   `group:"OIDC login config" namespace:"oidc" env-namespace:"OIDC"`
	Basic Basic  `group:"Basic login config" namespace:"basic" env-namespace:"BASIC"`
	DB    struct {
		URL          string        `long:"url" env:"URL" description:"Database URL" default:"sqlite://data.sqlite?cache=shared"`
		MaxConn      int           `long:"max-conn" env:"MAX_CONN" description:"Maximum number of opened connections to database" default:"10"`
		IdleConn     int           `long:"idle-conn" env:"IDLE_CONN" description:"Maximum number of idle connections to database" default:"1"`
		IdleTimeout  time.Duration `long:"idle-timeout" env:"IDLE_TIMEOUT" description:"Maximum amount of time a connection may be idle" default:"0"`
		ConnLifeTime time.Duration `long:"conn-life-time" env:"CONN_LIFE_TIME" description:"Maximum amount of time a connection may be reused" default:"0"`
	} `group:"Database configuration" namespace:"db" env-namespace:"DB"`
	Cache struct {
		Limit int           `long:"limit" env:"LIMIT" description:"Maximum number of tokens in cache" default:"1024"`
		TTL   time.Duration `long:"ttl" env:"TTL" description:"Maximum live time of token in cache" default:"15s"`
	} `group:"Cache configuration" namespace:"cache" env-namespace:"CACHE"`

	StatsInterval time.Duration `long:"stats-interval" env:"STATS_INTERVAL" description:"Interval of statistics synchronization" default:"1s"`
}

//nolint:maligned
type Server struct {
	Bind              string        `long:"bind" env:"BIND" description:"Bind address"`
	TLS               bool          `long:"tls" env:"TLS" description:"Enable TLS"`
	CA                string        `long:"ca" env:"CA" description:"Path to CA files. Optional unless IGNORE_SYSTEM_CA set" default:"ca.pem"`
	Cert              string        `long:"cert" env:"CERT" description:"Server certificate" default:"cert.pem"`
	Key               string        `long:"key" env:"KEY" description:"Server private key" default:"key.pem"`
	Mutual            bool          `long:"mutual" env:"MUTUAL" description:"Enable mutual TLS"`
	IgnoreSystemCA    bool          `long:"ignore-system-ca" env:"IGNORE_SYSTEM_CA" description:"Do not load system-wide CA"`
	ReadHeaderTimeout time.Duration `long:"read-header-timeout" env:"READ_HEADER_TIMEOUT" description:"How long to read header from the request" default:"3s"`
	Graceful          time.Duration `long:"graceful" env:"GRACEFUL" description:"Graceful shutdown timeout" default:"5s"`
}

type OIDC struct {
	ClientID     string `long:"client-id" env:"CLIENT_ID" description:"Client ID"`
	ClientSecret string `long:"client-secret" env:"CLIENT_SECRET" description:"Client secret"`
	Issuer       string `long:"issuer" env:"ISSUER" description:"OIDC issuer URL"`
	Session      string `long:"session" env:"SESSION" description:"Session storage" default:"local" choice:"local" choice:"redis"`
	Redis        struct {
		URL         string        `long:"url" env:"URL" description:"Redis URL" default:"redis://redis"`
		KeepAlive   time.Duration `long:"keep-alive" env:"KEEP_ALIVE" description:"Keep-alive interval" default:"30s"`
		Timeout     time.Duration `long:"timeout" env:"TIMEOUT" description:"Read/Write/Connect timeout" default:"5s"`
		MaxConn     int           `long:"max-conn" env:"MAX_CONN" description:"Maximum number of active connections" default:"10"`
		MaxIdle     int           `long:"max-idle" env:"MAX_IDLE" description:"Maximum number of idle connections" default:"1"`
		IdleTimeout time.Duration `long:"idle-timeout" env:"IDLE_TIMEOUT" description:"Close connections after remaining idle for this duration" default:"30s"`
	} `group:"OIDC Redis session configuration" namespace:"redis" env-namespace:"REDIS"`
	ServerURL string   `long:"server-url" env:"SERVER_URL" description:"(optional) public server URL for redirects"`
	Emails    []string `long:"emails" env:"EMAILS" description:"Allowed emails (enabled if at least one set)" env-delim:","`
}

type Basic struct {
	Realm    string `long:"realm" env:"REALM" description:"Realm name" default:"token-login"`
	User     string `long:"user" env:"USER" description:"User name" default:"admin"`
	Password string `long:"password" env:"PASSWORD" description:"User password hash from bcrypt" default:"$2y$05$d1BT6ay8qzViEGUjo4UDkOatWkFlszDfyzaXxCkM84kVhEJLtkXcu"` //  htpasswd -nbB user admin | cut -d ':' -f 2
}

func main() {
	var config Config
	config.Admin.Bind = ":8080"
	config.Auth.Bind = ":8081"

	parser := flags.NewParser(&config, flags.Default)
	parser.ShortDescription = "token-login"
	parser.LongDescription = fmt.Sprintf("Forward-auth server for tokens\ntoken-login %s, commit %s, built at %s by %s\nAuthor: Aleksandr Baryshnikov <owner@reddec.net>", version, commit, date, builtBy)

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := run(ctx, cancel, config); err != nil {
		panic(err)
	}
}

func run(ctx context.Context, cancel context.CancelFunc, config Config) error {
	// setup db
	store, err := config.getStore()
	if err != nil {
		return fmt.Errorf("create store: %w", err)
	}
	defer store.Close()

	// setup runners
	var wg multierror.Group

	// setup validation
	tokenValidator := validator.NewValidator(store, config.Cache.Limit, config.Cache.TTL)
	wg.Go(func() error {
		log.Println("starting stats dump every", config.StatsInterval)
		workerDumpStats(ctx, config, tokenValidator)
		return nil
	})
	// setup auth server
	wg.Go(func() error {
		defer cancel()
		router := chi.NewRouter()
		router.Get("/health", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNoContent)
		})
		router.Mount("/", web.AuthHandler(tokenValidator))
		return config.Auth.Run(ctx, cancel, "auth server", router)
	})
	// setup Admin server
	router := chi.NewRouter()
	router.Use(withOWASPHeaders)
	authMW := config.authMiddleware(ctx, router)
	router.With(authMW).Mount("/", web.NewAdmin(store))
	wg.Go(func() error {
		defer cancel()
		return config.Admin.Run(ctx, cancel, "admin server", router)
	})

	<-ctx.Done()
	cancel()
	return wg.Wait().ErrorOrNil()
}

func workerDumpStats(ctx context.Context, config Config, validator *validator.Validator) {
	t := time.NewTicker(config.StatsInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		if err := validator.UpdateStats(ctx); err != nil {
			log.Println("failed update stats:", err)
		}
	}
}

func (cfg Config) getStore() (dbo.Storage, error) { //nolint:ireturn
	u, err := url.Parse(cfg.DB.URL)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}

	switch u.Scheme {
	case "sqlite":
		return sqllite.New(cfg.DB.URL[len(u.Scheme)+3:], cfg.configurator())
	case "postgres":
		return pg.New(cfg.DB.URL, cfg.configurator())
	default:
		return nil, fmt.Errorf("unknown dialect %s", u.Scheme)
	}
}

func (cfg Config) configurator() func(db *sqlx.DB) {
	return func(db *sqlx.DB) {
		db.SetMaxIdleConns(cfg.DB.IdleConn)
		db.SetMaxOpenConns(cfg.DB.MaxConn)
		db.SetConnMaxIdleTime(cfg.DB.IdleTimeout)
		db.SetConnMaxLifetime(cfg.DB.ConnLifeTime)
	}
}

func (srv *Server) Run(ctx context.Context, cancel context.CancelFunc, name string, handler http.Handler) error {
	httpServer := &http.Server{
		Addr:              srv.Bind,
		Handler:           handler,
		ReadHeaderTimeout: srv.ReadHeaderTimeout,
	}

	tlsConfig, err := srv.tlsConfig()
	if err != nil {
		return fmt.Errorf("create TLS config: %w", err)
	}
	httpServer.TLSConfig = tlsConfig

	var wg multierror.Group

	wg.Go(func() error {
		defer cancel()
		var err error
		if srv.TLS {
			log.Println(name, "- starting TLS server on", srv.Bind)
			err = httpServer.ListenAndServeTLS(srv.Cert, srv.Key)
		} else {
			log.Println(name, "- starting plain HTTP server on", srv.Bind)
			err = httpServer.ListenAndServe()
		}
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return err
	})

	wg.Go(func() error {
		<-ctx.Done()
		log.Println(name, "- stopping")
		tctx, tcancel := context.WithTimeout(context.Background(), srv.Graceful)
		defer tcancel()
		return httpServer.Shutdown(tctx)
	})

	return wg.Wait().ErrorOrNil()
}

func (srv *Server) tlsConfig() (*tls.Config, error) {
	if !srv.TLS {
		return nil, nil //nolint:nilnil
	}
	var ca *x509.CertPool
	// create system-based CA or completely independent
	if srv.IgnoreSystemCA {
		ca = x509.NewCertPool()
	} else if roots, err := x509.SystemCertPool(); err == nil {
		ca = roots
	} else {
		return nil, fmt.Errorf("read system certs: %w", err)
	}

	// attach custom CA (if required)
	if err := srv.loadCA(ca); err != nil {
		return nil, fmt.Errorf("load CA: %w", err)
	}

	// read key
	cert, err := tls.LoadX509KeyPair(srv.Cert, srv.Key)
	if err != nil {
		return nil, fmt.Errorf("load cert and key: %w", err)
	}

	// enable mTLS if needed
	var clientAuth = tls.NoClientCert
	if srv.Mutual {
		clientAuth = tls.RequireAndVerifyClientCert
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		RootCAs:      ca,
		ClientCAs:    ca,
		ClientAuth:   clientAuth,
	}, nil
}

func (srv *Server) loadCA(ca *x509.CertPool) error {
	caCert, err := os.ReadFile(srv.CA)

	if err != nil {
		if srv.IgnoreSystemCA {
			// no system, no custom
			return fmt.Errorf("read CA: %w", err)
		}
		log.Println("failed read custom CA:", err)
		return nil
	}

	if !ca.AppendCertsFromPEM(caCert) {
		if srv.IgnoreSystemCA {
			return fmt.Errorf("CA certs failed to load")
		}
		log.Println("failed add custom CA to pool")
	}
	return nil
}

func (cfg Config) authMiddleware(ctx context.Context, router chi.Router) func(handler http.Handler) http.Handler {
	switch cfg.Login {
	case "basic":
		return cfg.Basic.createMiddleware(router)
	case "oidc":
		return cfg.OIDC.createMiddleware(ctx, router)
	default:
		panic("unknown login method " + cfg.Login)
	}
}

func (cfg *OIDC) emailsFilter() map[string]bool {
	var ans = make(map[string]bool, len(cfg.Emails))
	for _, e := range cfg.Emails {
		ans[strings.ToLower(e)] = true
	}
	return ans
}

func (cfg *OIDC) createMiddleware(ctx context.Context, router chi.Router) func(handler http.Handler) http.Handler {
	filter := cfg.emailsFilter()
	var session *scs.SessionManager
	if cfg.Session == "redis" {
		pool := &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return redis.DialURL(cfg.Redis.URL,
					redis.DialKeepAlive(cfg.Redis.KeepAlive),
					redis.DialWriteTimeout(cfg.Redis.Timeout),
					redis.DialReadTimeout(cfg.Redis.Timeout),
					redis.DialConnectTimeout(cfg.Redis.Timeout),
				)
			},
			MaxIdle:     cfg.Redis.MaxIdle,
			MaxActive:   cfg.Redis.MaxConn,
			IdleTimeout: cfg.Redis.IdleTimeout,
			Wait:        true,
		}
		session = scs.New()
		session.Store = redisstore.New(pool)
	}
	login, err := oidclogin.New(ctx, oidclogin.Config{
		IssuerURL:      cfg.Issuer,
		ClientID:       cfg.ClientID,
		ClientSecret:   cfg.ClientSecret,
		ServerURL:      cfg.ServerURL,
		SessionManager: session,
		PostAuth: func(writer http.ResponseWriter, req *http.Request, idToken *oidc.IDToken) error {
			if len(cfg.Emails) == 0 {
				return nil
			}
			email := oidclogin.Email(idToken)
			if !filter[strings.ToLower(email)] {
				return fmt.Errorf("email %s not allowed", email)
			}
			return nil
		},
	})
	if err != nil {
		panic(err)
	}
	router.Mount(oidclogin.Prefix, login)
	return func(handler http.Handler) http.Handler {
		return login.SecureFunc(func(writer http.ResponseWriter, request *http.Request) {
			token := oidclogin.Token(request)
			handler.ServeHTTP(writer, utils.WithUser(request, oidclogin.User(token)))
		})
	}
}

func (cfg *Basic) createMiddleware(router chi.Router) func(http.Handler) http.Handler {
	const flash = "_unauth"
	// mimic behaviour
	router.Get("/oauth/logout", func(writer http.ResponseWriter, request *http.Request) {
		utils.SetFlashPath(writer, flash, "true", "/") // potentially unsafe, but for logout should work fine
		writer.Header().Set("Location", "../")
		writer.WriteHeader(http.StatusSeeOther)
	})
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if utils.GetFlash(writer, request, flash) == "true" {
				writer.Header().Set("WWW-Authenticate", `Basic realm="`+cfg.Realm+`", charset="UTF-8"`)
				http.Error(writer, "Authorization required", http.StatusUnauthorized)
				return
			}
			user, password, ok := request.BasicAuth()
			if !ok ||
				subtle.ConstantTimeCompare([]byte(user), []byte(cfg.User)) == 0 ||
				bcrypt.CompareHashAndPassword([]byte(cfg.Password), []byte(password)) != nil {
				writer.Header().Set("WWW-Authenticate", `Basic realm="`+cfg.Realm+`", charset="UTF-8"`)
				http.Error(writer, "Authorization required", http.StatusUnauthorized)
				return
			}
			handler.ServeHTTP(writer, utils.WithUser(request, user))
		})
	}
}

func withOWASPHeaders(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		headers := writer.Header()
		headers.Set("X-Frame-Options", "DENY") // helps with click hijacking
		headers.Set("X-XSS-Protection", "1")
		headers.Set("X-Content-Type-Options", "nosniff")                  // helps with content-type substitution
		headers.Set("Referrer-Policy", "strict-origin-when-cross-origin") // disables cross-origin requests
		handler.ServeHTTP(writer, request)
	})
}
