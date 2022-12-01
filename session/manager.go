package session

import (
	"leason-toy-web/web"

	"github.com/google/uuid"
)

type Manager struct {
	Store
	Propagator
	CtxSessKey string
}

func (m *Manager) GetSession(ctx *web.Context) (Session, error) {
	if ctx.UserValues == nil {
		ctx.UserValues = make(map[string]interface{}, 1)
	}
	
	val, ok := ctx.UserValues[m.CtxSessKey]
	if ok {
		return val.(Session), nil
	}

	sessID, err := m.Propagator.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}

	sess, err := m.Store.Get(ctx.Req.Context(), sessID)
	if err != nil {
		return nil, err
	}

	ctx.UserValues[m.CtxSessKey] = sess

	return sess, nil
}

func (m *Manager) InitSession(ctx *web.Context) (Session, error) {
	sessID := uuid.New().String()
	sess, err := m.Store.Generate(ctx.Req.Context(), sessID)
	if err != nil {
		return nil, err
	}
	err = m.Inject(sessID, ctx.Resp)
	if err != nil {
		return nil, err
	}
	return sess, err
}

func (m *Manager) RemoveSession(ctx *web.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	m.Store.Remove(ctx.Req.Context(), sess.ID())
	if err != nil {
		return err
	}
	return m.Propagator.Remove(ctx.Resp)
}

func (m *Manager) RefreshSession(ctx *web.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	return m.Refresh(ctx.Req.Context(), sess.ID())
}
