//go:build e2e

package web

import (
	"html/template"
	template_engine "leason-toy-web/template"
	"mime/multipart"
	"path/filepath"
	"testing"
)

func TestFileUploader(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	if err != nil {
		t.Fatal(err)
	}
	s := NewHTTPServer(ServerWithTemplateEngine(&template_engine.GoTemplateEngine{T: tpl}))
	s.Get("/upload", func(ctx *Context) {
		er := ctx.Render("upload.gohtml", nil)
		if er != nil {
			t.Fatal(er)
		}
	})

	fu := FileUploader{
		FileField: "myfile",
		DstPathFunc: func(fileHeader *multipart.FileHeader) string {
			return filepath.Join("testdata", "upload", fileHeader.Filename)
		},
	}
	// 上传文件
	s.Post("/upload", fu.Handle())
	err = s.Start(":8081")
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileDownloader(t *testing.T) {
	s := NewHTTPServer()
	du := &FileDownloader{
		Dir: "./testdata/download",
	}
	s.Get("/download", du.Handle())
	// 在浏览器里面输入 localhost:8081/download?file=test.txt
	s.Start(":8081")
}

func TestStaticResourceHandler(t *testing.T) {
	s, err := NewStaticResourceHandler("testdata/static")
	if err != nil {
		t.Fatal(err)
	}

	server := NewHTTPServer()

	server.Get("/static/:filename", s.Handle)

	server.Start(":8081")
}
