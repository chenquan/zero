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
	"html/template"
	"net/http"
	"path"
	"strings"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(ctx *Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	*RouterGroup
	router        *router
	groups        []*RouterGroup     // 存储所有group
	htmlTemplates *template.Template // for html render
	funcMap       template.FuncMap   // for html render
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var middlewares []HandlerFunc
	// 获取分组中所有中间件
	for _, group := range e.groups {
		if strings.HasPrefix(request.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(writer, request)
	c.handlers = middlewares
	c.engine = e
	e.router.handle(c)
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}
func Default() *Engine {
	engine := New()
	engine.Use(LoggerDefault(), Recovery())
	return engine
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) LoadHTMLGlob(pattern string) {
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}
func (e *Engine) addRoute(method string, pattern string, handlers ...HandlerFunc) {
	e.router.addRoute(method, pattern, handlers...)
}
func (e *Engine) GET(pattern string, handlers ...HandlerFunc) {
	e.addRoute(GET, pattern, handlers...)
}
func (e *Engine) POST(pattern string, handlers ...HandlerFunc) {
	e.addRoute(POST, pattern, handlers...)
}
func (e *Engine) PUT(pattern string, handlers ...HandlerFunc) {
	e.addRoute(PUT, pattern, handlers...)
}
func (e *Engine) DELETE(pattern string, handlers ...HandlerFunc) {
	e.addRoute(DELETE, pattern, handlers...)
}
func (e *Engine) HEAD(pattern string, handlers ...HandlerFunc) {
	e.addRoute(HEAD, pattern, handlers...)
}

// Run 启动http服务器
func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

// 分组路由
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // 中间件
	parent      *RouterGroup  // 支持嵌套
	engine      *Engine       // 所有组共享一个Engine实例
}

func (rg *RouterGroup) Group(prefix string) *RouterGroup {

	newGroup := &RouterGroup{
		prefix: rg.prefix + prefix,
		parent: rg,
		engine: rg.engine,
	}
	rg.engine.groups = append(rg.engine.groups, newGroup)
	return newGroup
}

func (rg *RouterGroup) Use(middlewares ...HandlerFunc) {
	rg.middlewares = append(rg.middlewares, middlewares...)
}
func (rg *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(rg.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Res, c.Req)
	}
}

// serve static files
func (rg *RouterGroup) Static(relativePath string, root string) {
	handler := rg.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	rg.GET(urlPattern, handler)
}

// addRoute 添加路由
func (rg *RouterGroup) addRoute(method string, prefix string, handlers ...HandlerFunc) {
	// 真实路径 = 分组路径+当前路径
	pattern := rg.prefix + prefix
	rg.engine.addRoute(method, pattern, handlers...)
}

// GET 添加GET路由
func (rg *RouterGroup) GET(pattern string, handlers ...HandlerFunc) {
	rg.addRoute(GET, pattern, handlers...)
}

// POST 添加POST路由
func (rg *RouterGroup) POST(pattern string, handlers ...HandlerFunc) {
	rg.addRoute(POST, pattern, handlers...)
}

// PUT 添加PUT路由
func (rg *RouterGroup) PUT(pattern string, handlers ...HandlerFunc) {
	rg.addRoute(PUT, pattern, handlers...)
}

// DELETE 添加DELETE路由
func (rg *RouterGroup) DELETE(pattern string, handlers ...HandlerFunc) {
	rg.addRoute(POST, pattern, handlers...)
}

// HEAD 添加HEAD路由
func (rg *RouterGroup) HEAD(pattern string, handlers ...HandlerFunc) {
	rg.addRoute(HEAD, pattern, handlers...)
}
