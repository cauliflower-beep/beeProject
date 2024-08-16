package main

import (
	"beeWeb/day4-group-control/bee"
	"log"
	"net/http"
	"time"
)

func onlyForV2() bee.HandlerFunc {
	return func(c *bee.Context) {
		// Start timer
		t := time.Now()
		// if a server error occured
		// c.Fail(500, "Internal Server Error")
		// Calculate resolution time
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func main() {
	r := bee.New()
	r.Use(bee.Logger()) // global middleware
	/*
		handler 的参数变成了 bee.Context
		并且 bee.Context 封装了 HTML/String/JSON 函数，能够快速构造 HTTP 响应
	*/
	r.GET("/index", func(c *bee.Context) {
		c.HTML(http.StatusOK, "<h1>Index page</h1>")
	})

	v2 := r.Group("/v2")
	v2.Use(onlyForV2()) // v2 group middleware
	{
		v2.GET("/hello/:name", func(c *bee.Context) {
			// expect /hello/beeWeb
			c.String(http.StatusOK, "hello %s,you're at %s\n", c.Param("name"), c.Path)
		})
	}

	_ = r.Run(":9999")
}
