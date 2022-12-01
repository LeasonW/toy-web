package session

import (
	"leason-toy-web/web"
	"net/http"
	"testing"
)

func TestSession(t *testing.T) {
	var m Manager
	server := web.NewHTTPServer(web.ServerWithMiddleware(func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			if ctx.Req.URL.Path == "/login" {
				// 放过去，用户准备登陆
				next(ctx)
				return
			}

			_, err := m.GetSession(ctx)
			if err != nil {
				ctx.RespStatusCode = http.StatusUnauthorized
				ctx.RespData = []byte("请重新登陆")
				return
			}

			// 刷新 session 的过期时间
			_ = m.RefreshSession(ctx)
			next(ctx)

		}
	}))

	server.Get("/user", func(ctx *web.Context) {
		sess, _ := m.GetSession(ctx)
		sess.Get(ctx.Req.Context(), "nickname")
	})

	// 登陆
	server.Post("/login", func(ctx *web.Context) {
		// 要在这里校验登陆密码
		// ...

		sess, err := m.InitSession(ctx)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("登陆失败")
			return
		}
		err = sess.Set(ctx.Req.Context(), "nickname", "xiaoming")
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("登陆失败")
			return
		}
		ctx.RespStatusCode = http.StatusOK
		ctx.RespData = []byte("登陆成功")
	})

	// 登出
	server.Post("/logout", func(ctx *web.Context) {
		// 清理各种数据
		err := m.RemoveSession(ctx)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("登出失败")
			return
		}

		ctx.RespStatusCode = http.StatusOK
		ctx.RespData = []byte("登出成功")
	})

	server.Start(":8080")
}
