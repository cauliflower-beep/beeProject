package main

import (
	"fmt"
	"log"
	"net/http"
)

// Engine is the uni handler for all requests
type Engine struct{}

/*
空结构体Engine，实现了方法ServeHTTP, 可以作为 handler 传给 http.ListenAndServe
这个方法有2个参数，第二个参数是 Request ，该对象包含了该HTTP请求的所有的信息，比如请求地址、Header和Body等信息
第一个参数是 ResponseWriter ，利用 ResponseWriter 可以构造针对该请求的响应

至此，我们走出了实现Web框架的第一步，即，将所有的HTTP请求转向了我们自己的处理逻辑
比起 base1 ,我们拦截了所有的HTTP请求，拥有了统一的控制入口。
在这里我们可以自由定义路由映射的规则，也可以统一添加一些处理逻辑，例如日志、异常处理等。
*/
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		_, _ = fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	case "/hello":
		for k, v := range req.Header {
			_, _ = fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	default:
		_, _ = fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}

func main() {
	engine := new(Engine)
	log.Fatal(http.ListenAndServe(":9999", engine))
}
