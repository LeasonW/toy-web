package errhdl

import "leason-toy-web/web"

type MiddlewareBuilder struct {
	data map[int][]byte
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		data: make(map[int][]byte),
	}
}

func (m *MiddlewareBuilder) AddCode(code int, data []byte) *MiddlewareBuilder {
	m.data[code] = data
	return m
}

func (m *MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				if resp, ok := m.data[ctx.RespStatusCode]; ok {
					ctx.RespData = resp
				}
			}()
			next(ctx)
		}
	}
}
