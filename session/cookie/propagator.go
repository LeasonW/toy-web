package cookie

import "net/http"

type Propagator struct {
	cookieName   string
	cookieOption func(c *http.Cookie)
}

func NewPropagator() *Propagator {
	return &Propagator{
		cookieName:   "sessid",
		cookieOption: func(c *http.Cookie) {},
	}
}

type PropagatorOption func(p *Propagator)

func SetCookieOption(cookieOption func(c *http.Cookie)) PropagatorOption {
	return func(p *Propagator) {
		p.cookieOption = cookieOption
	}
}

func (p *Propagator) Inject(id string, resp http.ResponseWriter) error {
	c := &http.Cookie{
		Name:  p.cookieName,
		Value: id,
	}
	p.cookieOption(c)
	http.SetCookie(resp, c)
	return nil
}

func (p *Propagator) Extract(req *http.Request) (string, error) {
	c, err := req.Cookie(p.cookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func (p *Propagator) Remove(writer http.ResponseWriter) error {
	c := &http.Cookie{
		Name:   p.cookieName,
		MaxAge: -1,
	}
	http.SetCookie(writer, c)
	return nil
}
