package web

import (
	"fmt"
	"net/http"
	"testing"
)

// 这里放着端到端测试的代码

func TestServer(t *testing.T) {
	s := NewHTTPServer()
	s.Get("/", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, world"))
	})
	s.Get("/user", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, user"))
	})

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	s.Get("/cookie", func(ctx *Context) {
		ctx.SetCookie(&http.Cookie{Name: "sky", Value: "blue"})
		ctx.SetCookie(&http.Cookie{Name: "glass", Value: "green"})
		ctx.RespJSON(http.StatusOK, &User{Name: "zhanglixun", Age: 29})
	})

	s.Get("/print-cookie", func(ctx *Context) {
		cookies := make(map[string]string)
		for _, cookie := range ctx.Req.Cookies() {
			name := cookie.Name
			value := cookie.Value
			cookies[name] = value
		}
		ctx.RespJSON(http.StatusOK, cookies)
	})

	s.Start(":8081")
}

func Test_Route_Middleware(t *testing.T) {
	s := NewHTTPServer()
	s.Get("/", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, world"))
		fmt.Println("/ handler")
	})
	s.Get("/user", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, user"))
		fmt.Println("/user handler")
	})
	s.Use(http.MethodGet, "/", func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			fmt.Println("/ middleware before")
			next(ctx)
			fmt.Println("/ middleware after")
		}
	})
	s.Use(http.MethodGet, "/user", func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			fmt.Println("/user middleware before")
			next(ctx)
			fmt.Println("/user middleware after")
		}
	})

	s.Start(":8081")
}
