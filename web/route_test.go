package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func Test_router_AddRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		// 通配符测试用例
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		// 正则路由
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:name(^.+$)/abc",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path: "/",
				children: map[string]*node{
					"user": {
						path: "user",
						children: map[string]*node{
							"home": {path: "home", handler: mockHandler, typ: nodeTypeStatic},
						},
						handler: mockHandler,
						typ:     nodeTypeStatic,
					},
					"order": {
						path: "order",
						children: map[string]*node{
							"detail": {path: "detail", handler: mockHandler, typ: nodeTypeStatic},
						},
						starChild: &node{path: "*", handler: mockHandler, typ: nodeTypeAny},
						typ:       nodeTypeStatic,
					},
					"param": {
						path: "param",
						paramChild: &node{
							path:      ":id",
							paramName: "id",
							starChild: &node{
								path:    "*",
								handler: mockHandler,
								typ:     nodeTypeAny,
							},
							children: map[string]*node{"detail": {path: "detail", handler: mockHandler, typ: nodeTypeStatic}},
							handler:  mockHandler,
							typ:      nodeTypeParam,
						},
					},
				},
				starChild: &node{
					path: "*",
					children: map[string]*node{
						"abc": {
							path:      "abc",
							starChild: &node{path: "*", handler: mockHandler, typ: nodeTypeAny},
							handler:   mockHandler,
							typ:       nodeTypeStatic,
						},
					},
					starChild: &node{path: "*", handler: mockHandler, typ: nodeTypeAny},
					handler:   mockHandler,
					typ:       nodeTypeAny,
				},
				handler: mockHandler,
				typ:     nodeTypeStatic,
			},
			http.MethodPost: {
				path: "/",
				children: map[string]*node{
					"order": {path: "order", children: map[string]*node{
						"create": {path: "create", handler: mockHandler, typ: nodeTypeStatic},
					}},
					"login": {path: "login", handler: mockHandler, typ: nodeTypeStatic},
				},
				typ: nodeTypeStatic,
			},
			http.MethodDelete: {
				path: "/",
				children: map[string]*node{
					"reg": {
						path: "reg",
						typ:  nodeTypeStatic,
						regChild: &node{
							path:      ":id(.*)",
							paramName: "id",
							typ:       nodeTypeReg,
							handler:   mockHandler,
						},
					},
				},
				regChild: &node{
					path:      ":name(^.+$)",
					paramName: "name",
					typ:       nodeTypeReg,
					children: map[string]*node{
						"abc": {
							path:    "abc",
							handler: mockHandler,
						},
					},
				},
			},
		},
	}
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)

	// 非法用例
	r = newRouter()

	// 空字符串
	assert.PanicsWithValue(t, "web: 路由是空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})

	// 前导没有 /
	assert.PanicsWithValue(t, "web: 路由必须以 / 开头", func() {
		r.addRoute(http.MethodGet, "a/b/c", mockHandler)
	})

	// 后缀有 /
	assert.PanicsWithValue(t, "web: 路由不能以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	})

	// 根节点重复注册
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/]", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})
	// 普通节点重复注册
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/a/b/c]", func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})

	// 多个 /
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [/a//b]", func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [//a/b]", func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	})

	// 同时注册通配符路由，参数路由，正则路由
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [:id(.*)]", func() {
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/*", mockHandler)
		r.addRoute(http.MethodGet, "/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [*]", func() {
		r.addRoute(http.MethodGet, "/:id", mockHandler)
		r.addRoute(http.MethodGet, "/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [:id(.*)]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
	})
	// 参数冲突
	assert.PanicsWithValue(t, "web: 路由冲突，参数路由冲突，已有 :id，新注册 :name", func() {
		r.addRoute(http.MethodGet, "/a/b/c/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/c/:name", mockHandler)
	})
}

func (r router) equal(y router) (string, bool) {
	for k, v := range r.trees {
		yv, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("目标 router 里面没有方法 %s 的路由树", k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return k + "-" + str, ok
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if y == nil {
		return "目标节点为 nil", false
	}
	if n.path != y.path {
		return fmt.Sprintf("%s 节点 path 不相等 x %s, y %s", n.path, n.path, y.path), false
	}

	nhv := reflect.ValueOf(n.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("%s 节点 handler 不相等 x %s, y %s", n.path, nhv.Type().String(), yhv.Type().String()), false
	}

	if n.typ != y.typ {
		return fmt.Sprintf("%s 节点类型不相等 x %d, y %d", n.path, n.typ, y.typ), false
	}

	if n.paramName != y.paramName {
		return fmt.Sprintf("%s 节点参数名字不相等 x %s, y %s", n.path, n.paramName, y.paramName), false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("%s 子节点长度不等", n.path), false
	}
	if len(n.children) == 0 {
		return "", true
	}

	if n.starChild != nil {
		str, ok := n.starChild.equal(y.starChild)
		if !ok {
			return fmt.Sprintf("%s 通配符节点不匹配 %s", n.path, str), false
		}
	}
	if n.paramChild != nil {
		str, ok := n.paramChild.equal(y.paramChild)
		if !ok {
			return fmt.Sprintf("%s 路径参数节点不匹配 %s", n.path, str), false
		}
	}

	if n.regChild != nil {
		str, ok := n.regChild.equal(y.regChild)
		if !ok {
			return fmt.Sprintf("%s 路径参数节点不匹配 %s", n.path, str), false
		}
	}

	for k, v := range n.children {
		yv, ok := y.children[k]
		if !ok {
			return fmt.Sprintf("%s 目标节点缺少子节点 %s", n.path, k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return n.path + "-" + str, ok
		}
	}
	return "", true
}

func Test_router_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},

		// 正则
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:id([0-9]+)/home",
		},
	}

	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name   string
		method string
		path   string
		found  bool
		mi     *matchInfo
	}{
		{
			name:   "method not found",
			method: http.MethodHead,
		},
		{
			name:   "path not found",
			method: http.MethodGet,
			path:   "/abc",
		},
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "/",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "user",
			method: http.MethodGet,
			path:   "/user",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "user",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "no handler",
			method: http.MethodPost,
			path:   "/order",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path: "order",
				},
			},
		},
		{
			name:   "two layer",
			method: http.MethodPost,
			path:   "/order/create",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "create",
					handler: mockHandler,
				},
			},
		},
		// 通配符匹配
		{
			// 命中/order/*
			name:   "star match",
			method: http.MethodPost,
			path:   "/order/delete",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中通配符在中间的
			// /user/*/home
			name:   "star in middle",
			method: http.MethodGet,
			path:   "/user/Tom/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			// 比 /order/* 多了一段
			name:   "overflow",
			method: http.MethodPost,
			path:   "/order/delete/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		// 参数匹配
		{
			// 命中 /param/:id
			name:   ":id",
			method: http.MethodGet,
			path:   "/param/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/*
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/abc",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/detail
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/detail",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "detail",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /reg/:id(.*)
			name:   ":id(.*)",
			method: http.MethodDelete,
			path:   "/reg/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /:id([0-9]+)/home
			name:   ":id([0-9]+)",
			method: http.MethodDelete,
			path:   "/123/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 未命中 /:id([0-9]+)/home
			name:   "not :id([0-9]+)",
			method: http.MethodDelete,
			path:   "/abc/home",
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			assert.Equal(t, tc.mi.pathParams, mi.pathParams)
			n := mi.n
			wantVal := reflect.ValueOf(tc.mi.n.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}

func Test_isRegExpr(t *testing.T) {
	name, isReg := isRegExpr(":id(.*)")
	assert.True(t, isReg)
	assert.True(t, name == "id")

	name, isReg = isRegExpr(":name(^.+$)")
	assert.True(t, isReg)
	assert.True(t, name == "name")

	name, isReg = isRegExpr("*")
	assert.False(t, isReg)
	assert.True(t, name == "")

	name, isReg = isRegExpr(":user")
	assert.False(t, isReg)
	assert.True(t, name == "")

	name, isReg = isRegExpr("user")
	assert.False(t, isReg)
	assert.True(t, name == "")
}

// CASE 1
// Use("GET", "/a/b", ms)
// 当输入路径 /a/b 的时候，会调度对应的 ms
// 当输入路径 /a/b/c 的时候，会调度执行 ms
func Test_FindRoute_Case1(t *testing.T) {
	mockHandler := HandleFunc(func(ctx *Context) {})
	ms := Middleware(func(next HandleFunc) HandleFunc { return func(ctx *Context) {} })

	testRoutes := []struct {
		method  string
		path    string
		mdls    []Middleware
		handler func(ctx *Context)
	}{
		{
			method:  http.MethodGet,
			path:    "/a/b",
			mdls:    []Middleware{ms},
			handler: mockHandler,
		},
		{
			method:  http.MethodGet,
			path:    "/a/b/c",
			handler: mockHandler,
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, tr.handler, tr.mdls...)
	}

	// 执行 /a/b
	match, ok := r.findRoute(http.MethodGet, "/a/b")
	assert.True(t, ok)
	assert.NotNil(t, match)
	assert.NotNil(t, match.n)
	assert.True(t, len(match.mdls) == 1)
	assert.Equal(t, reflect.ValueOf(match.mdls[0]), reflect.ValueOf(ms))
	assert.Equal(t, reflect.ValueOf(match.n.handler), reflect.ValueOf(mockHandler))
	// 执行 /a/b/c
	match, ok = r.findRoute(http.MethodGet, "/a/b")
	assert.True(t, ok)
	assert.NotNil(t, match)
	assert.NotNil(t, match.n)
	assert.True(t, len(match.mdls) == 1)
	assert.Equal(t, reflect.ValueOf(match.mdls[0]), reflect.ValueOf(ms))
	assert.Equal(t, reflect.ValueOf(match.n.handler), reflect.ValueOf(mockHandler))
}

// CASE 2
// Use(GET, "/a", ms1)
// Use(GET, "/a/b", ms2)
// 当输入路径 /a/c 的时候，会调度 ms1
// 当输入路径为 /a/b/c 的时候，会调度 ms1 和 ms2
func Test_FindRoute_Case2(t *testing.T) {
	mockHandler := HandleFunc(func(ctx *Context) {})
	ms1 := Middleware(func(next HandleFunc) HandleFunc { return func(ctx *Context) {} })
	ms2 := Middleware(func(next HandleFunc) HandleFunc { return func(ctx *Context) {} })

	testRoutes := []struct {
		method  string
		path    string
		mdls    []Middleware
		handler func(ctx *Context)
	}{
		{
			method:  http.MethodGet,
			path:    "/a/c",
			handler: mockHandler,
		},
		{
			method:  http.MethodGet,
			path:    "/a/b/c",
			handler: mockHandler,
		},
		{
			method: http.MethodGet,
			path:   "/a",
			mdls:   []Middleware{ms1},
		},
		{
			method: http.MethodGet,
			path:   "/a/b",
			mdls:   []Middleware{ms2},
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, tr.handler, tr.mdls...)
	}

	// 执行 /a/c，预期调度 ms1
	match, ok := r.findRoute(http.MethodGet, "/a/c")
	assert.True(t, ok)
	assert.NotNil(t, match)
	assert.NotNil(t, match.n)
	assert.Equal(t, reflect.ValueOf(match.n.handler), reflect.ValueOf(mockHandler))
	assert.True(t, len(match.mdls) == 1)
	assert.Equal(t, reflect.ValueOf(match.mdls[0]), reflect.ValueOf(ms1))

	// 执行 /a/b/c，预期调度 ms1 -> ms2
	match, ok = r.findRoute(http.MethodGet, "/a/b/c")
	assert.True(t, ok)
	assert.NotNil(t, match)
	assert.NotNil(t, match.n)
	assert.Equal(t, reflect.ValueOf(match.n.handler), reflect.ValueOf(mockHandler))
	assert.True(t, len(match.mdls) == 2)
	assert.Equal(t, reflect.ValueOf(match.mdls[0]), reflect.ValueOf(ms1))
	assert.Equal(t, reflect.ValueOf(match.mdls[1]), reflect.ValueOf(ms2))
}

// CASE 3
// Use(GET, "/a/*/c", ms1)
// Use(GET, "/a/b/c", ms2)
// 当输入路径 /a/d/c 的时候，会调度 ms1
// 当输入路径为 /a/b/c 的时候，会调度 ms1 和 ms2
// 当输入路径为 /a/b/d 的时候，不会调度
func Test_FindRoute_Case3(t *testing.T) {
	mockHandler := HandleFunc(func(ctx *Context) {})
	ms1 := Middleware(func(next HandleFunc) HandleFunc { return func(ctx *Context) {} })
	ms2 := Middleware(func(next HandleFunc) HandleFunc { return func(ctx *Context) {} })

	testRoutes := []struct {
		method  string
		path    string
		mdls    []Middleware
		handler func(ctx *Context)
	}{
		{
			method:  http.MethodGet,
			path:    "/a/d/c",
			handler: mockHandler,
		},
		{
			method:  http.MethodGet,
			path:    "/a/b/c",
			handler: mockHandler,
		},
		{
			method:  http.MethodGet,
			path:    "/a/b/d",
			handler: mockHandler,
		},
		{
			method: http.MethodGet,
			path:   "/a/*/c",
			mdls:   []Middleware{ms1},
		},
		{
			method: http.MethodGet,
			path:   "/a/b/c",
			mdls:   []Middleware{ms2},
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, tr.handler, tr.mdls...)
	}

	// 执行 /a/d/c，预期调度 ms1
	match, ok := r.findRoute(http.MethodGet, "/a/d/c")
	assert.True(t, ok)
	assert.NotNil(t, match)
	assert.NotNil(t, match.n)
	assert.Equal(t, reflect.ValueOf(match.n.handler), reflect.ValueOf(mockHandler))
	assert.True(t, len(match.mdls) == 1)
	assert.Equal(t, reflect.ValueOf(match.mdls[0]), reflect.ValueOf(ms1))

	// 执行 /a/b/c，预期调度 ms1 -> ms2
	match, ok = r.findRoute(http.MethodGet, "/a/b/c")
	assert.True(t, ok)
	assert.NotNil(t, match)
	assert.NotNil(t, match.n)
	assert.Equal(t, reflect.ValueOf(match.n.handler), reflect.ValueOf(mockHandler))
	assert.True(t, len(match.mdls) == 2)
	assert.Equal(t, reflect.ValueOf(match.mdls[0]), reflect.ValueOf(ms1))
	assert.Equal(t, reflect.ValueOf(match.mdls[1]), reflect.ValueOf(ms2))

	// 执行 /a/b/d，预期不会调度
	match, ok = r.findRoute(http.MethodGet, "/a/b/d")
	fmt.Printf("match = %#v\n", match)
	assert.True(t, ok)
	assert.NotNil(t, match)
	assert.NotNil(t, match.n)
	assert.Equal(t, reflect.ValueOf(match.n.handler), reflect.ValueOf(mockHandler))
	assert.True(t, len(match.mdls) == 0)
}

// CASE 4
// Use(GET, "/a/:id", ms1)
// Use(GET, "/a/123/c", ms2)
// 当输入路径为 /a/123 的时候，会调度 ms1
// 当输入路径为 /a/123/c 的时候，会调度 ms1 和 ms2
func Test_FindRoute_Case4(t *testing.T) {
	mockHandler := HandleFunc(func(ctx *Context) {})
	ms1 := Middleware(func(next HandleFunc) HandleFunc { return func(ctx *Context) {} })
	ms2 := Middleware(func(next HandleFunc) HandleFunc { return func(ctx *Context) {} })

	testRoutes := []struct {
		method  string
		path    string
		mdls    []Middleware
		handler func(ctx *Context)
	}{
		{
			method:  http.MethodGet,
			path:    "/a/123",
			handler: mockHandler,
		},
		{
			method:  http.MethodGet,
			path:    "/a/123/c",
			handler: mockHandler,
		},
		{
			method: http.MethodGet,
			path:   "/a/:id",
			mdls:   []Middleware{ms1},
		},
		{
			method: http.MethodGet,
			path:   "/a/123/c",
			mdls:   []Middleware{ms2},
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, tr.handler, tr.mdls...)
	}

	// 执行 /a/123, 预期调度 ms1
	match, ok := r.findRoute(http.MethodGet, "/a/123")
	assert.True(t, ok)
	assert.NotNil(t, match)
	assert.NotNil(t, match.n)
	assert.Equal(t, reflect.ValueOf(match.n.handler), reflect.ValueOf(mockHandler))
	assert.True(t, len(match.mdls) == 1)
	assert.Equal(t, reflect.ValueOf(match.mdls[0]), reflect.ValueOf(ms1))

	// 执行 /a/123/c，预期调度 ms1 -> ms2
	match, ok = r.findRoute(http.MethodGet, "/a/123/c")
	assert.True(t, ok)
	assert.NotNil(t, match)
	assert.NotNil(t, match.n)
	assert.Equal(t, reflect.ValueOf(match.n.handler), reflect.ValueOf(mockHandler))
	assert.True(t, len(match.mdls) == 2)
	assert.Equal(t, reflect.ValueOf(match.mdls[0]), reflect.ValueOf(ms1))
	assert.Equal(t, reflect.ValueOf(match.mdls[1]), reflect.ValueOf(ms2))
}
