package template

import (
	"bytes"
	"context"
	"html/template"
)

type GoTemplateEngine struct {
	T *template.Template
}

func (g *GoTemplateEngine) Render(ctx context.Context, tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := g.T.ExecuteTemplate(bs, tplName, data)
	if err != nil {
		return nil, err
	}
	return bs.Bytes(), nil
}
