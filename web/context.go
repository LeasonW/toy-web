package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type Context struct {
	Req        *http.Request
	Resp       http.ResponseWriter
	PathParams map[string]string
	Route      string

	RespData       []byte
	RespStatusCode int

	cacheQueryValues url.Values
}

func (c *Context) BindJSON(val interface{}) error {
	if c.Req.Body == nil {
		return errors.New("web: body is nil")
	}
	jsonDecoder := json.NewDecoder(c.Req.Body)
	return jsonDecoder.Decode(val)
}

func (c *Context) FormValue(key string) (string, error) {
	if err := c.Req.ParseForm(); err != nil {
		return "", err
	}
	return c.Req.FormValue(key), nil
}

func (c *Context) QueryValue(key string) (string, error) {
	if c.cacheQueryValues == nil {
		c.cacheQueryValues = c.Req.URL.Query()
	}

	vals, ok := c.cacheQueryValues[key]
	if !ok {
		return "", errors.New("web: key not exist")
	}
	return vals[0], nil
}

func (c *Context) PathValue(key string) (string, error) {
	val, ok := c.PathParams[key]
	if !ok {
		return "", errors.New("web: key not exist")
	}
	return val, nil
}

func (c *Context) RespJSON(code int, val interface{}) error {
	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}

	c.RespStatusCode = code
	c.RespData = bs
	return nil
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp, cookie)
}
