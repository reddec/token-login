package server

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"

	"github.com/reddec/token-login/api"
	"github.com/reddec/token-login/internal/ent"
	"github.com/reddec/token-login/internal/ent/project"
	"github.com/reddec/token-login/internal/ent/token"
	"github.com/reddec/token-login/internal/types"
	"github.com/reddec/token-login/internal/utils"
)

type (
	UpdateHandler func(id int)
	RemoveHandler func(id int)
)

func New(client *ent.Client) *Server {
	return &Server{client: client}
}

type Server struct {
	client   *ent.Client
	onUpdate []UpdateHandler
	onRemove []RemoveHandler
}

func (srv *Server) OnUpdate(fn UpdateHandler) {
	srv.onUpdate = append(srv.onUpdate, fn)
}

func (srv *Server) OnRemove(fn RemoveHandler) {
	srv.onRemove = append(srv.onRemove, fn)
}

func (srv *Server) CreateToken(ctx context.Context, req *api.TokenConfig) (*api.Credential, error) {
	key, err := types.NewKey()
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	headers := parseHeaders(req.Headers)
	// validate config
	_, err = types.NewAccessKey(key.Hash(), req.Host.Value, req.Path.Value)
	if err != nil {
		return nil, fmt.Errorf("validate key: %w", err)
	}

	user := utils.GetUser(ctx)
	kid := key.ID()

	// Validate project belongs to user (if specified)
	if req.ProjectId != 0 {
		exists, err := srv.client.Project.Query().
			Where(project.ID(req.ProjectId), project.User(user)).
			Exist(ctx)
		if err != nil {
			return nil, fmt.Errorf("check project: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("project %d not found", req.ProjectId)
		}
	}

	t, err := srv.client.Token.Create().
		SetUser(user).
		SetHash(key.Hash()).
		SetKeyID(&kid).
		SetLabel(req.Label.Value).
		SetHeaders(headers).
		SetHost(req.Host.Value).
		SetPath(req.Path.Value).
		SetProjectID(req.ProjectId).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}
	srv.notifyUpdated(t.ID)
	return &api.Credential{
		ID:  t.ID,
		Key: key.String(),
	}, nil
}

func (srv *Server) DeleteToken(ctx context.Context, params api.DeleteTokenParams) error {
	removed, err := srv.client.Token.Delete().Where(
		token.User(utils.GetUser(ctx)),
		token.ID(params.Token),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete token: %w", err)
	}
	if removed > 0 {
		srv.notifyRemoved(params.Token)
	}
	return nil
}

func (srv *Server) GetToken(ctx context.Context, params api.GetTokenParams) (*api.Token, error) {
	t, err := srv.client.Token.Query().Where(
		token.User(utils.GetUser(ctx)),
		token.ID(params.Token),
	).WithProject().Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	return mapToken(t), nil
}

func (srv *Server) ListTokens(ctx context.Context, params api.ListTokensParams) ([]api.Token, error) {
	q := srv.client.Token.Query().Where(token.User(utils.GetUser(ctx))).WithProject().Order(token.ByID(sql.OrderDesc()))
	if p, ok := params.Project.Get(); ok {
		q = q.Where(token.ProjectID(p))
	}
	list, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}
	var out = make([]api.Token, 0, len(list))
	for _, t := range list {
		x := mapToken(t)
		out = append(out, *x)
	}
	return out, nil
}

func (srv *Server) RefreshToken(ctx context.Context, params api.RefreshTokenParams) (*api.Credential, error) {
	key, err := types.NewKey()

	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	kid := key.ID()

	changed, err := srv.client.Token.Update().
		Where(
			token.User(utils.GetUser(ctx)),
			token.ID(params.Token),
		).
		SetHash(key.Hash()).
		SetKeyID(&kid).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("update token: %w", err)
	}
	if changed == 0 {
		return nil, errors.New("unknown token")
	}
	srv.notifyUpdated(params.Token)
	return &api.Credential{
		ID:  params.Token,
		Key: key.String(),
	}, nil
}

func (srv *Server) UpdateToken(ctx context.Context, req *api.TokenPatch, params api.UpdateTokenParams) error {
	upd := srv.client.Token.Update().Where(
		token.User(utils.GetUser(ctx)),
		token.ID(params.Token),
	)

	if req.Host.Set {
		upd.SetHost(req.Host.Value)
	}

	if req.Path.Set {
		upd.SetPath(req.Path.Value)
	}

	if req.Label.Set {
		upd.SetLabel(req.Label.Value)
	}

	if req.Headers != nil {
		upd.SetHeaders(parseHeaders(req.Headers))
	}

	changed, err := upd.Save(ctx)
	if err != nil {
		return fmt.Errorf("update token: %w", err)
	}
	if changed == 0 {
		return errors.New("unknown token")
	}
	srv.notifyUpdated(params.Token)
	return nil
}

func (srv *Server) notifyUpdated(id int) {
	for _, h := range srv.onUpdate {
		h(id)
	}
}

func (srv *Server) notifyRemoved(id int) {
	for _, h := range srv.onRemove {
		h(id)
	}
}

func parseHeaders(v []api.NameValue) types.Headers {
	out := make(types.Headers, 0, len(v))
	for _, it := range v {
		out = append(out, types.Header{
			Name:  it.Name,
			Value: it.Value,
		})
	}
	return out
}

func mapToken(t *ent.Token) *api.Token {
	projectID := 0
	projectSlug := ""
	if t.Edges.Project != nil {
		projectID = t.Edges.Project.ID
		projectSlug = t.Edges.Project.Slug
	}
	return &api.Token{
		ID:        t.ID,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		LastAccessAt: api.OptDateTime{
			Value: t.LastAccessAt,
			Set:   !t.LastAccessAt.IsZero(),
		},
		KeyID:       t.KeyID.String(),
		User:        t.User,
		Label:       t.Label,
		Host:        t.Host,
		Path:        t.Path,
		Headers:     mapHeaders(t.Headers),
		Requests:    t.Requests,
		ProjectId:   projectID,
		ProjectSlug: projectSlug,
	}
}

func mapHeaders(v types.Headers) []api.NameValue {
	out := make([]api.NameValue, 0, len(v))
	for _, p := range v {
		out = append(out, api.NameValue{
			Name:  p.Name,
			Value: p.Value,
		})
	}
	return out
}

// --- Project CRUD ---

func (srv *Server) ListProjects(ctx context.Context) ([]api.Project, error) {
	list, err := srv.client.Project.Query().
		Where(project.User(utils.GetUser(ctx))).
		Order(project.ByID(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	out := make([]api.Project, 0, len(list))
	for _, p := range list {
		out = append(out, *mapProject(p))
	}
	return out, nil
}

func (srv *Server) CreateProject(ctx context.Context, req *api.ProjectConfig) (*api.Project, error) {
	builder := srv.client.Project.Create().
		SetUser(utils.GetUser(ctx)).
		SetSlug(req.Slug)
	if desc, ok := req.Description.Get(); ok {
		builder.SetDescription(desc)
	}
	p, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return mapProject(p), nil
}

func (srv *Server) GetProject(ctx context.Context, params api.GetProjectParams) (*api.Project, error) {
	p, err := srv.client.Project.Query().
		Where(
			project.ID(params.Project),
			project.User(utils.GetUser(ctx)),
		).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	return mapProject(p), nil
}

func (srv *Server) UpdateProject(ctx context.Context, req *api.ProjectPatch, params api.UpdateProjectParams) error {
	upd := srv.client.Project.Update().Where(
		project.ID(params.Project),
		project.User(utils.GetUser(ctx)),
	)
	if desc, ok := req.Description.Get(); ok {
		upd.SetDescription(desc)
	}
	changed, err := upd.Save(ctx)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	if changed == 0 {
		return errors.New("unknown project")
	}
	return nil
}

func (srv *Server) DeleteProject(ctx context.Context, params api.DeleteProjectParams) error {
	// Protect the default project (empty slug) from deletion
	user := utils.GetUser(ctx)
	p, err := srv.client.Project.Query().
		Where(
			project.ID(params.Project),
			project.User(user),
		).Only(ctx)
	if err != nil {
		return fmt.Errorf("get project: %w", err)
	}
	if p.Slug == "" {
		return errors.New("cannot delete default project")
	}

	// Collect affected token IDs for cache invalidation
	tokenIDs, err := srv.client.Token.Query().
		Where(token.ProjectID(params.Project)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("list tokens in project: %w", err)
	}

	// Unlink tokens from this project
	if _, err := srv.client.Token.Update().
		Where(token.ProjectID(params.Project)).
		ClearProjectID().
		Save(ctx); err != nil {
		return fmt.Errorf("unlink tokens from project: %w", err)
	}

	if err := srv.client.Project.DeleteOneID(params.Project).Exec(ctx); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	// Invalidate cache for all affected tokens
	for _, tid := range tokenIDs {
		srv.notifyUpdated(tid)
	}
	return nil
}

func mapProject(p *ent.Project) *api.Project {
	return &api.Project{
		ID:          p.ID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		Slug:        p.Slug,
		Description: p.Description,
	}
}
