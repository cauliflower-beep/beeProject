package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	/*
		此时, 我们调用 http.HandleFunc 实现了路由和Handler的映射，但只能针对具体的路由写处理逻辑
	*/
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":9999", nil)) // 启动web服务 param1表示监听的端口 param2表示处理所有的http请求实例
}

// indexHandler handler echoes r.URL.Path
/*
	go中，客户端请求信息都封装到了 Request 对象
	但是发送给客户端的响应并不是 Response 对象，而是 ResponseWriter
	实际上，在底层支撑 ResponseWriter 的结构体就是 http.Response
	详见 net/http 包下 server.go 中的 readRequest 方法(调用处理器处理 HTTP 请求时调用了该方法返回响应对象)，
	并且其返回值是 response 指针，这也是为什么在处理器方法声明的时候 Request 是指针类型，而 ResponseWriter 不是，
	实际上在底层，响应对象也是指针类型(因为在应用代码中需要设置响应头和响应实体，所以响应对象理应是指针类型).
	response 结构体定义和 ResponseWriter 一样都位于 server.go，
	不过由于 response 对外不可见，所以只能通过 ResponseWriter 接口访问它。
	两者之间的关系是 ResponseWriter 是一个接口，而 http.response 实现了它。当我们引用 ResponseWriter 时，实际上引用的是 http.response 对象实例。
*/
func indexHandler(w http.ResponseWriter, req *http.Request) {
	/*
		Fprintf 把格式字符串输出到指定文件设备中，主要用于文件操作，格式化输出到一个stream 通常是到文件
	*/
	fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
}

// helloHandler echoes r.URL.Header
func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
}
