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
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	engine := Default()
	engine.Use(func(ctx *Context) {
		fmt.Println("第一层")
		ctx.Next()
	})
	v1 := engine.Group("/test")
	v1.Use(func(ctx *Context) {
		fmt.Println("第一层-v1")
		ctx.Next()

	})
	{
		v1.GET("/test", func(ctx *Context) {
			fmt.Println("第一层-v1-test")
			//context.Status(200)
			ctx.JSON(200, "hhhhh")
		})
	}
	engine.Run(":8080")

}
