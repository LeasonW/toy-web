package accesslog

import (
	"fmt"
	web "leason-toy-web/web"
	"net/http"
	"testing"
)

func Test_AccessLog_Middleware(t *testing.T) {
	mdl := NewMiddlewareBuilder().LogFunc(
		func(val string) {
			fmt.Println(val)
		}).Build()

	server := web.NewHTTPServer(web.ServerWithMiddleware(mdl))

	server.Get("/a/*", func(ctx *web.Context) {
		ctx.RespJSON(http.StatusOK, "hello, it's me")
	})

	server.Start(":8081")
}
