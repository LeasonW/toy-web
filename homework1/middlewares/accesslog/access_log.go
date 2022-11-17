package accesslog

import (
	"encoding/json"
	"fmt"
	web "leason-toy-web/homework1"
)

type MiddlewareBuilder struct {
	logFn func(val string)
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

func (m *MiddlewareBuilder) LogFunc(fn func(val string)) *MiddlewareBuilder {
	m.logFn = fn
	return m
}

func (m *MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(c *web.Context) {
			defer func() {
				l := accessLog{
					Host:       c.Req.Host,
					HTTPMethod: c.Req.Method,
					Path:       c.Req.URL.Path,
				}
				data, _ := json.Marshal(l)
				m.logFn(string(data))
			}()
			next(c)
		}
	}
}

func AccessLogMiddlerwarefunc(next web.HandleFunc) web.HandleFunc {
	return func(c *web.Context) {
		defer func() {
			l := accessLog{
				Host:       c.Req.Host,
				HTTPMethod: c.Req.Method,
				Path:       c.Req.URL.Path,
			}
			data, _ := json.Marshal(l)
			fmt.Println(string(data))
		}()
		next(c)
	}
}

type accessLog struct {
	Host       string `json:"host,omitempty"`
	Route      string `json:"route,omitempty"`
	HTTPMethod string `json:"http_method,omitempty"`
	Path       string `json:"path,omitempty"`
}
