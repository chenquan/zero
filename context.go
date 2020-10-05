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
	"sync"
	"time"
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

	// 该互斥锁保护Keys map
	mu sync.RWMutex

	// keys 是专门针对每个请求的上下文的键/值对
	keys map[string]interface{}
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

/****************************
 * 上下文数据存储
 ****************************/
// Set 设置键和值
func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.keys == nil {
		c.keys = make(map[string]interface{})
	}

	c.keys[key] = value
}

// Get 返回与键关联的接口值
func (c *Context) Get(key string) (value interface{}, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok = c.keys[key]
	return
}

// MustGet 返回与键关联的接口值,如果不存在则报错
func (c *Context) MustGet(key string) interface{} {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString 返回与键关联的字符串
func (c *Context) GetString(key string) (s string, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		s, ok = val.(string)
	}
	return
}

// GetBool 返回与键关联的布尔值
func (c *Context) GetBool(key string) (b bool, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, ok = val.(bool)
	}
	return
}

// GetInt 返回与值关联的整型值
func (c *Context) GetInt(key string) (i int, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		i, ok = val.(int)
	}
	return
}

// GetInt64 返回与值关联的64位整型值
func (c *Context) GetInt64(key string) (i64 int64, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, ok = val.(int64)
	}
	return
}

// GetUint 返回与值关联的整型值
func (c *Context) GetUint(key string) (ui uint, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		ui, ok = val.(uint)
	}
	return
}

// GetUint64 返回与值关联的无符号64位整型值
func (c *Context) GetUint64(key string) (ui64 uint64, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		ui64, ok = val.(uint64)
	}
	return
}

// GetFloat64 返回与值关联的双精度浮点数
func (c *Context) GetFloat64(key string) (f64 float64, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, ok = val.(float64)
	}
	return
}

// GetTime 返回与键关联的时间值
func (c *Context) GetTime(key string) (t time.Time, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		t, ok = val.(time.Time)
	}
	return
}

// GetDuration 返回与键关联的时间间隔值.
func (c *Context) GetDuration(key string) (d time.Duration, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		d, ok = val.(time.Duration)
	}
	return
}

// GetStringSlice 返回与键关联的切片字符串值.
func (c *Context) GetStringSlice(key string) (ss []string, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		ss, ok = val.([]string)
	}
	return
}

// GetStringMap 返回与键关联的值作为接口映射.
func (c *Context) GetStringMap(key string) (sm map[string]interface{}, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		sm, ok = val.(map[string]interface{})
	}
	return
}

// GetStringMapString 返回与键关联的值作为字符串映射.
func (c *Context) GetStringMapString(key string) (sms map[string]string, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		sms, ok = val.(map[string]string)
	}
	return
}

// GetStringMapStringSlice 返回与键关联的值，作为映射到字符串切片的映射.
func (c *Context) GetStringMapStringSlice(key string) (smss map[string][]string, ok bool) {
	if val, ok := c.Get(key); ok && val != nil {
		smss, ok = val.(map[string][]string)
	}
	return
}
