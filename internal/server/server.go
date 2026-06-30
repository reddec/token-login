package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/reddec/token-login/api"
	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/internal/types"
	"github.com/reddec/token-login/internal/utils"
)

type (
	UpdateHandler func(id int)
	RemoveHandler func(id int)
)

func New(store dbo.Store) *Server {
	return &Server{store: store}
}

type Server struct {
	store    dbo.Store
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
	_, err = types.NewAccessKey(key.Hash(), req.Host.Value, req.Path.Value)
	if err != nil {
		return nil, fmt.Errorf("validate key: %w", err)
	}

	user := utils.GetUser(ctx)
	kid := key.ID()

	if req.ProjectId != 0 {
		exists, err := srv.store.ProjectExists(ctx, user, int64(req.ProjectId))
		if err != nil {
			return nil, fmt.Errorf("check project: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("project %d not found", req.ProjectId)
		}
	}

	t, err := srv.store.CreateToken(ctx, dbo.CreateTokenParams{
		User:      user,
		Hash:      key.Hash(),
		KeyID:     &kid,
		Label:     req.Label.Value,
		Headers:   headers,
		Host:      req.Host.Value,
		Path:      req.Path.Value,
		ProjectID: int64(req.ProjectId),
	})
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}
	srv.notifyUpdated(int(t.ID))
	return &api.Credential{
		ID:  int(t.ID),
		Key: key.String(),
	}, nil
}

func (srv *Server) DeleteToken(ctx context.Context, params api.DeleteTokenParams) error {
	removed, err := srv.store.DeleteToken(ctx, utils.GetUser(ctx), int64(params.Token))
	if err != nil {
		return fmt.Errorf("delete token: %w", err)
	}
	if removed > 0 {
		srv.notifyRemoved(params.Token)
	}
	return nil
}

func (srv *Server) GetToken(ctx context.Context, params api.GetTokenParams) (*api.Token, error) {
	t, err := srv.store.GetToken(ctx, utils.GetUser(ctx), int64(params.Token))
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	return mapToken(t), nil
}

func (srv *Server) ListTokens(ctx context.Context, params api.ListTokensParams) ([]api.Token, error) {
	var projectID int64
	if p, ok := params.Project.Get(); ok {
		projectID = int64(p)
	}
	list, err := srv.store.ListTokens(ctx, utils.GetUser(ctx), projectID)
	if err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}
	out := make([]api.Token, 0, len(list))
	for _, t := range list {
		out = append(out, *mapToken(t))
	}
	return out, nil
}

func (srv *Server) RefreshToken(ctx context.Context, params api.RefreshTokenParams) (*api.Credential, error) {
	key, err := types.NewKey()
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	kid := key.ID()

	changed, err := srv.store.RefreshToken(ctx, utils.GetUser(ctx), int64(params.Token), key.Hash(), &kid)
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
	p := dbo.UpdateTokenParams{
		User: utils.GetUser(ctx),
		ID:   int64(params.Token),
	}
	if req.Host.Set {
		p.Host = &req.Host.Value
	}
	if req.Path.Set {
		p.Path = &req.Path.Value
	}
	if req.Label.Set {
		p.Label = &req.Label.Value
	}
	if req.Headers != nil {
		h := parseHeaders(req.Headers)
		p.Headers = &h
	}

	changed, err := srv.store.UpdateToken(ctx, p)
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

func (srv *Server) ListProjects(ctx context.Context) ([]api.Project, error) {
	list, err := srv.store.ListProjects(ctx, utils.GetUser(ctx))
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
	p, err := srv.store.CreateProject(ctx, dbo.CreateProjectParams{
		User:        utils.GetUser(ctx),
		Slug:        req.Slug,
		Description: req.Description.Or(""),
	})
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return mapProject(p), nil
}

func (srv *Server) GetProject(ctx context.Context, params api.GetProjectParams) (*api.Project, error) {
	p, err := srv.store.GetProject(ctx, utils.GetUser(ctx), int64(params.Project))
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	return mapProject(p), nil
}

func (srv *Server) UpdateProject(ctx context.Context, req *api.ProjectPatch, params api.UpdateProjectParams) error {
	changed, err := srv.store.UpdateProject(ctx, dbo.UpdateProjectParams{
		User:        utils.GetUser(ctx),
		ID:          int64(params.Project),
		Description: req.Description.Or(""),
	})
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	if changed == 0 {
		return errors.New("unknown project")
	}
	return nil
}

func (srv *Server) DeleteProject(ctx context.Context, params api.DeleteProjectParams) error {
	user := utils.GetUser(ctx)
	p, err := srv.store.GetProject(ctx, user, int64(params.Project))
	if err != nil {
		return fmt.Errorf("get project: %w", err)
	}
	if p.Slug == "" {
		return errors.New("cannot delete default project")
	}

	tokenIDs, err := srv.store.DeleteProject(ctx, user, int64(params.Project))
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	for _, tid := range tokenIDs {
		srv.notifyUpdated(int(tid))
	}
	return nil
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

func mapToken(t *dbo.Token) *api.Token {
	return &api.Token{
		ID:        int(t.ID),
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
		ProjectId:   int(t.ProjectID),
		ProjectSlug: t.ProjectSlug,
	}
}

func mapProject(p *dbo.Project) *api.Project {
	return &api.Project{
		ID:          int(p.ID),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		Slug:        p.Slug,
		Description: p.Description,
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
