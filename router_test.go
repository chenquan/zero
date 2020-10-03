/*
 *
 *    Copyright 2020. Chen Quan
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
	"fmt"
	"testing"
)

func newTestRouter() *router {
	r := newRouter()
	//r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/assets/*filepath", nil)
	r.addRoute("GET", "/test/:id/:name", nil)
	return r
}

func Test_parsePattern(t *testing.T) {
	var s []string
	for i, ss := range s {
		fmt.Println(i, ss)
	}
}

func Test_router_getRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/hello/chenquan")

	if n == nil {
		t.Fatal("nil shouldn't be returned")
	}

	if n.pattern != "/hello/:name" {
		t.Fatal("should match /hello/:name")
	}

	if ps["name"] != "chenquan" {
		t.Fatal("name should be equal to 'chenquan'")
	}
	fmt.Printf("matched path: %s, params['name']: %s\n", n.pattern, ps["name"])
	n, ps = r.getRoute("GET", "/test/11/chenquan")
	if ps["id"] != "11" {
		t.Fatal("name should be equal to 'chenquan'")
	}
	if ps["name"] != "chenquan" {
		t.Fatal("name should be equal to 'chenquan'")
	}
	fmt.Printf("matched path: %s, params['id']: %s and ['name']: %s\n", n.pattern, ps["id"], ps["name"])

}

func Test_router_handle(t *testing.T) {

}
