package main

import (
	"beeWeb/day4-group-control/bee"
	"net/http"
)

func main() {
	r := bee.New()
	/*
		handler 的参数变成了 bee.Context
		并且 bee.Context 封装了 HTML/String/JSON 函数，能够快速构造 HTTP 响应
	*/
	r.GET("/index", func(c *bee.Context) {
		c.HTML(http.StatusOK, "<h1>Index page</h1>")
	})
	v1 := r.Group("/v1")
	{
		v1.GET("/index", func(c *bee.Context) {
			c.HTML(http.StatusOK, "<h1>Hello bee</h1>")
		})

		v1.GET("/hello", func(c *bee.Context) {
			// expect  /hello?name=goku
			c.String(http.StatusOK, "hello %s,you're at %s\n", c.Query("name"), c.Path)
		})
	}

	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *bee.Context) {
			// expect /hello/beeWeb
			c.String(http.StatusOK, "hello %s,you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(ctx *bee.Context) {
			ctx.JSON(http.StatusOK, bee.H{
				"username": ctx.PostForm("username"),
				"password": ctx.PostForm("password"),
			})
		})
	}

	_ = r.Run(":9999")
}
