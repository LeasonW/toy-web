package session

import (
	"context"
	"net/http"
)

// Store 定义接口管理session本身
type Store interface {
	// Generate
	// session ID 由谁来指定
	// 要不要在接口维度上设置超时时间
	// 要不要在 Store 内部生成ID
	Generate(ctx context.Context, id string) (Session, error)
	Refresh(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (Session, error)
}

type Session interface {
	Get(ctx context.Context, key string) (any interface{}, err error)
	Set(ctx context.Context, key string, value interface{}) error
	ID() string
}

// Propagator 是一个抽象层，不同的实现允许将 session id 存储在不同的地方（只是主流存储在 Cookie 而已）
type Propagator interface {
	Inject(id string, resp http.ResponseWriter) error
	Extract(req *http.Request) (string, error)
	Remove(writer http.ResponseWriter) error
}
