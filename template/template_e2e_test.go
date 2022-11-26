//go build:e2e
package template

import (
	"html/template"
	"leason-toy-web/web"
	"testing"
)

func TestServerWithRenderEngine(t *testing.T) {
	tpl, err := template.ParseGlob("../testdata/tpls/*.gohtml")
	if err != nil {
		t.Fatal(err)
	}
	s := web.NewHTTPServer(web.ServerWithTemplateEngine(&GoTemplateEngine{T: tpl}))
	s.Get("/login", func(ctx *web.Context) {
		er := ctx.Render("login.gohtml", nil)
		if er != nil {
			t.Fatal(er)
		}
	})
	err = s.Start(":8081")
	if err != nil {
		t.Fatal(err)
	}
}
