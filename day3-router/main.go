package main

import (
	"beeWeb/day3-router/bee"
	"net/http"
)

func main() {
	r := bee.New()
	/*
		handler 的参数变成了 bee.Context
		并且 bee.Context 封装了 HTML/String/JSON 函数，能够快速构造 HTTP 响应
	*/
	r.GET("/", func(c *bee.Context) {
		c.HTML(http.StatusOK, "<h1>hello bee</h1>")
	})
	r.GET("/hello", func(c *bee.Context) {
		// expect  /hello?name=goku
		c.String(http.StatusOK, "hello %s,you're at %s\n", c.Query("name"), c.Path)
	})

	r.GET("/hello/:name", func(c *bee.Context) {
		// expect /hello/beeWeb
		c.String(http.StatusOK, "hello %s,you're at %s\n", c.Query("name"), c.Path)
	})

	r.GET("/assets/*filepath", func(c *bee.Context) {
		c.JSON(http.StatusOK, bee.H{"filepath": c.Param("filepath")})
	})

	_ = r.Run(":9999")
}
