package web

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	regPattern = regexp.MustCompile(`^:(.+)\((.*)\)`)
)

type router struct {
	// trees 是按照 HTTP 方法来组织的
	// 如 GET => *node
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// addRoute 注册路由。
// method 是 HTTP 方法
// - 已经注册了的路由，无法被覆盖。例如 /user/home 注册两次，会冲突
// - path 必须以 / 开始并且结尾不能有 /，中间也不允许有连续的 /
// - 不能在同一个位置注册不同的参数路由，例如 /user/:id 和 /user/:name 冲突
// - 不能在同一个位置同时注册通配符路由和参数路由，例如 /user/:id 和 /user/* 冲突
// - 同名路径参数，在路由匹配的时候，值会被覆盖。例如 /user/:id/abc/:id，那么 /user/123/abc/456 最终 id = 456
func (r *router) addRoute(method string, path string, handler HandleFunc, mdls ...Middleware) {
	r.validatePath(path)

	root, ok := r.trees[method]
	if !ok {
		root = &node{
			path:  "/",
			route: "/",
			typ:   nodeTypeStatic,
		}
		r.trees[method] = root
	}

	if path == "/" {
		if handler != nil {
			if root.handler != nil {
				panic("web: 路由冲突[/]")
			}
			root.handler = handler
		}
		root.mdls = append(root.mdls, mdls...)
		return
	}

	segs := strings.Split(path, "/")
	for _, seg := range segs[1:] {
		if seg == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		root = root.childOrCreate(seg)
	}

	if handler != nil {
		if root.handler != nil {
			panic(fmt.Sprintf("web: 路由冲突[%s]", path))
		}
		root.handler = handler
	}

	root.route = path
	for _, mdl := range mdls {
		root.mdls = append(root.mdls, mdl)
	}
}

// findRoute 查找对应的节点
// 注意，返回的 node 内部 HandleFunc 不为 nil 才算是注册了路由
func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return &matchInfo{
			n:    root,
			mdls: root.mdls,
		}, true
	}

	var params map[string]string
	var anyNode *node
	segs := strings.Split(path, "/")
	for _, seg := range segs[1:] {
		child, ok := root.childOf(seg)
		if !ok {
			if anyNode != nil {
				return &matchInfo{
					n:          anyNode,
					pathParams: params,
					mdls:       r.findMdls(r.trees[method], segs[1:]),
				}, true
			}
			return nil, false
		}
		if child.typ == nodeTypeAny {
			anyNode = child
		}
		if child.typ == nodeTypeReg || child.typ == nodeTypeParam {
			if params == nil {
				params = make(map[string]string)
			}
			params[child.paramName] = seg
		}
		root = child
	}
	return &matchInfo{
		n:          root,
		pathParams: params,
		mdls:       r.findMdls(r.trees[method], segs[1:]),
	}, true
}

type nodeType int

const (
	// 静态路由
	nodeTypeStatic = iota
	// 正则路由
	nodeTypeReg
	// 路径参数路由
	nodeTypeParam
	// 通配符路由
	nodeTypeAny
)

// node 代表路由树的节点
// 路由树的匹配顺序是：
// 1. 静态完全匹配
// 2. 正则匹配，形式 :param_name(reg_expr)
// 3. 路径参数匹配：形式 :param_name
// 4. 通配符匹配：*
// 这是不回溯匹配
type node struct {
	typ nodeType

	route string
	path  string
	// children 子节点
	// 子节点的 path => node
	children map[string]*node
	// handler 命中路由之后执行的逻辑
	handler HandleFunc

	// 通配符 * 表达的节点，任意匹配
	starChild *node

	paramChild *node
	// 正则路由和参数路由都会使用这个字段
	paramName string

	// 正则表达式
	regChild *node
	regExpr  *regexp.Regexp

	// middleware
	mdls []Middleware
}

// child 返回子节点
// 第一个返回值 *node 是命中的节点
// 第二个返回值 bool 代表是否命中
func (n *node) childOf(path string) (*node, bool) {
	// 优先匹配静态节点
	child, ok := n.children[path]
	if ok {
		return child, true
	}

	// 其次匹配正则节点
	if n.regChild != nil {
		if matchRegExpr(n.regChild.path, path) {
			return n.regChild, true
		}
	}
	// 再次匹配参数节点
	if n.paramChild != nil {
		return n.paramChild, true
	}

	// 最后匹配通配符节点
	return n.starChild, n.starChild != nil
}

// childOrCreate 查找子节点，
// 首先会判断 path 是不是通配符路径
// 其次判断 path 是不是参数路径，即以 : 开头的路径
// 最后会从 children 里面查找，
// 如果没有找到，那么会创建一个新的节点，并且保存在 node 里面
func (n *node) childOrCreate(path string) *node {
	if path == "*" {
		if n.regChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [%s]", path))
		}
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}
		if n.starChild == nil {
			n.starChild = &node{
				typ:  nodeTypeAny,
				path: path,
			}
		}
		return n.starChild
	}

	if name, isReg := isRegExpr(path); isReg {
		if n.starChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [%s]", path))
		}
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [%s]", path))
		}
		if n.regChild == nil {
			n.regChild = &node{
				typ:       nodeTypeReg,
				path:      path,
				paramName: name,
			}
		}
		return n.regChild
	}

	if string(path[0]) == ":" {
		if n.starChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}
		if n.regChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [%s]", path))
		}

		if n.paramChild == nil {
			n.paramChild = &node{
				typ:       nodeTypeParam,
				path:      path,
				paramName: path[1:],
			}
		} else if n.paramChild.path != path {
			panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.paramChild.path, path))
		}
		return n.paramChild
	}

	if n.children == nil {
		n.children = make(map[string]*node)
	}

	child, ok := n.children[path]
	if !ok {
		child = &node{
			typ:  nodeTypeStatic,
			path: path,
		}
		n.children[path] = child
	}
	return child
}

// findMdls 层次遍历所有满足前缀匹配的节点
// 按照 通配符匹配 -> 路径匹配 -> 正则匹配 -> 静态匹配的优先级，返回节点上的 middlewares
func (r *router) findMdls(root *node, segs []string) []Middleware {
	mdls := []Middleware{}
	mdls = append(mdls, root.mdls...)
	queue := []*node{root}

	for _, seg := range segs {
		if len(queue) == 0 {
			break
		}

		l := len(queue)
		for i := 0; i < l; i++ {
			n := queue[i]
			if n.starChild != nil {
				mdls = append(mdls, n.starChild.mdls...)
				queue = append(queue, n.starChild)
			}
			if n.paramChild != nil {
				mdls = append(mdls, n.paramChild.mdls...)
				queue = append(queue, n.paramChild)
			}
			if n.regChild != nil && matchRegExpr(n.regChild.path, seg) {
				mdls = append(mdls, n.regChild.mdls...)
				queue = append(queue, n.regChild)
			}
			if n.children != nil {
				if child, ok := n.children[seg]; ok {
					mdls = append(mdls, child.mdls...)
					queue = append(queue, child)
				}
			}
		}
		queue = queue[l:]
	}
	return mdls
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
	mdls       []Middleware
}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		// 大多数情况，参数路径只会有一段
		m.pathParams = map[string]string{key: value}
	}
	m.pathParams[key] = value
}

func isRegExpr(path string) (string, bool) {
	match := regPattern.FindStringSubmatch(path)
	if len(match) != 3 {
		return "", false
	}

	return match[1], true
}

func matchRegExpr(pattern string, path string) bool {
	match := regPattern.FindStringSubmatch(pattern)
	if len(match) != 3 {
		return false
	}

	reg := regexp.MustCompile(match[2])
	match = reg.FindStringSubmatch(path)
	return len(match) > 0
}

func (r *router) validatePath(path string) {
	if path == "" {
		panic("web: 路由是空字符串")
	}
	if string(path[0]) != "/" {
		panic("web: 路由必须以 / 开头")
	}
	if len(path) > 1 && string(path[len(path)-1]) == "/" {
		panic("web: 路由不能以 / 结尾")
	}
}
