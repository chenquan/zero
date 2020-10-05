/*
 *
 *    Copyright 2020 Chen Quan
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package zero

import (
	"log"
	"reflect"
	"runtime"
	"strings"
)

// 支持的请求方式
const (
	GET     = "GET"
	HEAD    = "HEAD"
	POST    = "POST"
	PUT     = "PUT"
	PATCH   = "PATCH" // RFC 5789
	DELETE  = "DELETE"
	CONNECT = "CONNECT"
	OPTIONS = "OPTIONS"
	TRACE   = "TRACE"
)

// router 路由
type router struct {
	roots    map[string]*node // key是网络请求方式
	handlers map[string][]HandlerFunc
}

// newRouter 新建路由
func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string][]HandlerFunc),
	}
}

// parsePattern 解析Pattern
// 只允许一个*存在
func parsePattern(pattern string) []string {
	// <path0>/<path1>/<path2>/*
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

// addRoute 添加一个新的路由
func (r *router) addRoute(method string, pattern string, handlers ...HandlerFunc) {
	var fns []string
	for _, handler := range handlers {
		fn := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		fns = append(fns, fn)
	}
	handlerNames := ""
	if len(fns) != 0 {

		handlerNames = ": " + strings.Join(fns, ", ")
	}
	log.Printf("[ZERO] ROUTE %4s - %s, handlers(%d)%s", method, pattern, len(handlers), handlerNames)

	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		// 路由根节点
		r.roots[method] = new(node)
	}
	// 插入路径
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handlers
}

// getRoute 获取路由节点和路径参数
func (r *router) getRoute(method string, path string) (node *node, params map[string]string) {
	root, ok := r.roots[method]
	if !ok {
		return
	}

	searchParts := parsePattern(path)
	// 存储路径参数
	params = make(map[string]string)

	// 获取路由节点
	node = root.search(searchParts, 0)

	if node != nil {
		parts := parsePattern(node.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return
	}
	return
}

func (r *router) handle(c *Context) {
	// 获取路由
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers, r.handlers[key]...)

	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			panic(NotFoundError)
		})
	}
	c.Next()

}
