package main

import (
	"beeWeb/day2-context/bee"
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

	r.POST("/login", func(c *bee.Context) {
		c.JSON(http.StatusOK, bee.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})

	_ = r.Run(":9999")
}
