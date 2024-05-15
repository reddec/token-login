package main

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gomodule/redigo/redis"
	"github.com/hashicorp/go-multierror"
	"github.com/jessevdk/go-flags"
	oidclogin "github.com/reddec/oidc-login"
	"golang.org/x/crypto/bcrypt"

	"github.com/reddec/token-login/api"
	"github.com/reddec/token-login/internal/cache"
	"github.com/reddec/token-login/internal/ent"
	"github.com/reddec/token-login/internal/plumbing"
	"github.com/reddec/token-login/internal/server"
	"github.com/reddec/token-login/internal/utils"

	"github.com/reddec/token-login/web"
)

//nolint:gochecknoglobals
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

type Config struct {
	Admin Server    `group:"Admin server configuration" namespace:"admin" env-namespace:"ADMIN"`
	Auth  Server    `group:"Auth server configuration" namespace:"auth" env-namespace:"AUTH"`
	Login string    `long:"login" env:"LOGIN" description:"Login method for admin UI" default:"basic" choice:"basic" choice:"oidc" choice:"proxy"`
	OIDC  OIDC      `group:"OIDC login config" namespace:"oidc" env-namespace:"OIDC"`
	Basic Basic     `group:"Basic login config" namespace:"basic" env-namespace:"BASIC"`
	Proxy ProxyAuth `group:"Proxy login config" namespace:"proxy" env-namespace:"PROXY"`
	DB    struct {
		URL          string        `long:"url" env:"URL" description:"Database URL" default:"sqlite://data.sqlite?cache=shared&_fk=1&_pragma=foreign_keys(1)"`
		MaxConn      int           `long:"max-conn" env:"MAX_CONN" description:"Maximum number of opened connections to database" default:"10"`
		IdleConn     int           `long:"idle-conn" env:"IDLE_CONN" description:"Maximum number of idle connections to database" default:"1"`
		IdleTimeout  time.Duration `long:"idle-timeout" env:"IDLE_TIMEOUT" description:"Maximum amount of time a connection may be idle" default:"0"`
		ConnLifeTime time.Duration `long:"conn-life-time" env:"CONN_LIFE_TIME" description:"Maximum amount of time a connection may be reused" default:"0"`
	} `group:"Database configuration" namespace:"db" env-namespace:"DB"`
	Cache struct {
		TTL time.Duration `long:"ttl" env:"TTL" description:"Maximum live time of token in cache. Also forceful reload time" default:"15s"`
	} `group:"Cache configuration" namespace:"cache" env-namespace:"CACHE"`
	Stats struct {
		Buffer   int           `long:"buffer" env:"BUFFER" description:"Buffer size for hits" default:"2048"`
		Interval time.Duration `long:"interval" env:"INTERVAL" description:"Statistics interval" default:"5s"`
	} `group:"Stats configuration" namespace:"stats" env-namespace:"STATS"`
	Debug struct {
		Enable      bool   `long:"enable" env:"ENABLE" description:"Enable debug mode"`
		Impersonate string `long:"impersonate" env:"IMPERSONATE" description:"Disable normal auth and use static user name"`
	} `group:"Debug" namespace:"debug" env-namespace:"DEBUG"`
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
	config.setupLogging()

	// setup db
	store, err := ent.New(ctx, config.DB.URL, config.configureDatabase)
	if err != nil {
		return fmt.Errorf("create store: %w", err)
	}
	defer store.Close()

	hitsCache := make(chan web.Hit, config.Stats.Buffer)
	keysCache := cache.New(store)

	if err := keysCache.SyncKeys(ctx); err != nil {
		// initial sync
		return fmt.Errorf("sync keys: %w", err)
	}
	srv := server.New(store)
	apiServer, err := api.NewServer(srv)
	if err != nil {
		return fmt.Errorf("create api server: %w", err)
	}
	srv.OnRemove(keysCache.Drop)
	srv.OnUpdate(func(id int) {
		if err := keysCache.SyncKey(ctx, id); err != nil {
			slog.Error("sync key failed", "id", id, "err", err)
		}
	})

	// setup runners
	var wg multierror.Group

	// setup db->cache key sync
	wg.Go(func() error {
		defer cancel()
		keysCache.PollKeys(ctx, config.Cache.TTL)
		return nil
	})

	// setup stats->db sync
	wg.Go(func() error {
		defer cancel()
		plumbing.SyncStats(ctx, store, hitsCache, config.Stats.Interval)
		return nil
	})

	// setup auth server
	wg.Go(func() error {
		defer cancel()
		router := chi.NewRouter()
		router.Get("/health", func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusNoContent)
		})
		router.Mount("/", web.AuthHandler(keysCache, hitsCache))
		return config.Auth.Run(ctx, cancel, "auth server", router)
	})

	// setup Admin server
	router := chi.NewRouter()
	if config.Debug.Enable {
		router.Use(func(handler http.Handler) http.Handler {
			// enable logging
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				started := time.Now()
				handler.ServeHTTP(w, r)
				dur := time.Since(started)
				slog.Debug("http request complete", "method", r.Method, "path", r.URL.Path, "duration", dur)
			})
		})
		router.Use(cors.AllowAll().Handler)
	} else {
		router.Use(withOWASPHeaders)
	}
	authMW := config.authMiddleware(ctx, router)

	router.With(authMW).Route("/", func(r chi.Router) {
		r.Mount(api.Prefix+"/", http.StripPrefix(api.Prefix, apiServer))
		r.Mount("/", http.FileServerFS(web.Assets()))
	})

	wg.Go(func() error {
		defer cancel()
		return config.Admin.Run(ctx, cancel, "admin server", router)
	})
	slog.Info("ready", "version", version, "debug", config.Debug.Enable)
	<-ctx.Done()
	cancel()
	return wg.Wait().ErrorOrNil()
}

func (cfg Config) configureDatabase(db *sql.DB) {
	db.SetMaxIdleConns(cfg.DB.IdleConn)
	db.SetMaxOpenConns(cfg.DB.MaxConn)
	db.SetConnMaxIdleTime(cfg.DB.IdleTimeout)
	db.SetConnMaxLifetime(cfg.DB.ConnLifeTime)
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
			slog.Info("starting TLS server", "bind", srv.Bind, "name", name)
			err = httpServer.ListenAndServeTLS(srv.Cert, srv.Key)
		} else {
			slog.Info("starting plain server", "bind", srv.Bind, "name", name)
			err = httpServer.ListenAndServe()
		}
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return err
	})

	wg.Go(func() error {
		<-ctx.Done()
		slog.Info("stopping server", "name", name)
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
		slog.Warn("failed read custom CA", "error", err)
		return nil
	}

	if !ca.AppendCertsFromPEM(caCert) {
		if srv.IgnoreSystemCA {
			return errors.New("CA certs failed to load")
		}
		slog.Warn("failed add custom CA to pool")
	}
	return nil
}

func (cfg Config) authMiddleware(ctx context.Context, router chi.Router) func(handler http.Handler) http.Handler {
	if cfg.Debug.Impersonate != "" {
		slog.Warn("Authorization disabled", "user", cfg.Debug.Impersonate)
		return (&NoAuth{User: cfg.Debug.Impersonate}).createMiddleware(router)
	}
	switch cfg.Login {
	case "basic":
		return cfg.Basic.createMiddleware(router)
	case "oidc":
		return cfg.OIDC.createMiddleware(ctx, router)
	case "proxy":
		return cfg.Proxy.createMiddleware(router)
	default:
		panic("unknown login method " + cfg.Login)
	}
}

func (cfg Config) setupLogging() {
	if !cfg.Debug.Enable {
		return
	}
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))
	slog.SetDefault(logger)
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
		PostAuth: func(_ http.ResponseWriter, _ *http.Request, idToken *oidc.IDToken) error {
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
			request = request.WithContext(utils.WithUser(request.Context(), oidclogin.User(token)))
			handler.ServeHTTP(writer, request)
		})
	}
}

func (cfg *Basic) createMiddleware(router chi.Router) func(http.Handler) http.Handler {
	const flash = "_unauth"
	// mimic behaviour
	router.Get("/oauth/logout", func(writer http.ResponseWriter, _ *http.Request) {
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
			request = request.WithContext(utils.WithUser(request.Context(), user))
			handler.ServeHTTP(writer, request)
		})
	}
}

type ProxyAuth struct {
	Header string `long:"header" env:"HEADER" description:"Header which will contain user name" default:"X-User"`
	Logout string `long:"logout" env:"LOGOUT" description:"Logout redirect"`
}

func (pa *ProxyAuth) createMiddleware(router chi.Router) func(http.Handler) http.Handler {
	router.Get("/oauth/logout", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Location", pa.Logout)
		writer.WriteHeader(http.StatusSeeOther)
	})
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			request = request.WithContext(utils.WithUser(request.Context(), request.Header.Get(pa.Header)))
			handler.ServeHTTP(writer, request)
		})
	}
}

type NoAuth struct {
	User string
}

func (na *NoAuth) createMiddleware(router chi.Router) func(http.Handler) http.Handler {
	router.Get("/oauth/logout", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Location", "")
		writer.WriteHeader(http.StatusSeeOther)
	})
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			request = request.WithContext(utils.WithUser(request.Context(), na.User))
			handler.ServeHTTP(writer, request)
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
