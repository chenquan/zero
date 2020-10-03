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
	"encoding/json"
	"fmt"
	"net/http"
)

type Z map[string]interface{}

type Context struct {
	// request info
	Req        *http.Request     // 请求
	Path       string            //路径
	Method     string            // 请求方式
	Params     map[string]string // 路径参数
	StatusCode int
	// response info
	Res http.ResponseWriter //返回

	// middleware
	handlers []HandlerFunc // 中间件
	index    int           //当前执行的中间件,-1:初始位置
	engine   *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Res:    w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

// Next 执行下一个中间件
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}
func (c *Context) Query(key string) {
	c.Req.URL.Query().Get(key)
}
func (c *Context) Status(statusCode int) {
	c.StatusCode = statusCode
	c.Res.WriteHeader(statusCode)
}
func (c *Context) SetHeader(key string, value string) {
	c.Res.Header().Set(key, value)
}
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Res.Write([]byte(fmt.Sprintf(format, values...)))
}
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Res)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Res, err.Error(), 500)
	}
}
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Res.Write(data)
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// Fail 错误信息
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, Z{"message": err})
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Res, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
