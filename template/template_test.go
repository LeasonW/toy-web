package template

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	type User struct {
		Name string
	}

	tpl := template.New("hello-world")
	tpl, err := tpl.Parse("Hello, {{.Name}}")
	if err != nil {
		t.Fatal(err)
	}

	bs := &bytes.Buffer{}
	err = tpl.Execute(bs, User{
		Name: "Tom",
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "Hello, Tom", bs.String())
}

func TestMapData(t *testing.T) {
	tpl := template.New("map-data")
	tpl, err := tpl.Parse("Hello, {{.Name}}")
	require.NoError(t, err)

	bs := &bytes.Buffer{}
	err = tpl.Execute(bs, map[string]string{
		"Name": "Tom",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello, Tom", bs.String())
}

func TestSliceData(t *testing.T) {
	tpl := template.New("slice-data")
	tpl, err := tpl.Parse("Hello, {{index . 0}}")
	require.NoError(t, err)

	bs := &bytes.Buffer{}
	err = tpl.Execute(bs, []string{"Tom"})
	require.NoError(t, err)
	assert.Equal(t, "Hello, Tom", bs.String())
}

func TestFuncCall(t *testing.T) {
	tpl := template.New("func-call")
	tpl, err := tpl.Parse(`
切片长度: {{len .Slice}}
say hello: {{.Hello "Tom" "Jerry"}}
打印数字: {{printf "%.2f" 1.234}}
`)
	require.NoError(t, err)

	bs := &bytes.Buffer{}
	err = tpl.Execute(bs, FuncCall{
		Slice: []string{"Tom", "Jerry"},
	})
	require.NoError(t, err)
	assert.Equal(t, `
切片长度: 2
say hello: Tom-Jerry
打印数字: 1.23
`, bs.String())
}

type FuncCall struct {
	Slice []string
}

func (f FuncCall) Hello(first string, last string) string {
	return fmt.Sprintf("%s-%s", first, last)
}

func TestLoop(t *testing.T) {
	tpl := template.New("loop")
	tpl, err := tpl.Parse(`
{{- range $idx, $elem := . -}}
下标: {{- $idx -}}
{{- end -}}
`)
	require.NoError(t, err)

	bs := &bytes.Buffer{}
	err = tpl.Execute(bs, make([]bool, 100))
	require.NoError(t, err)
	assert.Equal(t, `下标:0下标:1下标:2下标:3下标:4下标:5下标:6下标:7下标:8下标:9下标:10下标:11下标:12下标:13下标:14下标:15下标:16下标:17下标:18下标:19下标:20下标:21下标:22下标:23下标:24下标:25下标:26下标:27下标:28下标:29下标:30下标:31下标:32下标:33下标:34下标:35下标:36下标:37下标:38下标:39下标:40下标:41下标:42下标:43下标:44下标:45下标:46下标:47下标:48下标:49下标:50下标:51下标:52下标:53下标:54下标:55下标:56下标:57下标:58下标:59下标:60下标:61下标:62下标:63下标:64下标:65下标:66下标:67下标:68下标:69下标:70下标:71下标:72下标:73下标:74下标:75下标:76下标:77下标:78下标:79下标:80下标:81下标:82下标:83下标:84下标:85下标:86下标:87下标:88下标:89下标:90下标:91下标:92下标:93下标:94下标:95下标:96下标:97下标:98下标:99`, bs.String())
}
