package bee

import (
	"net/http"
)

// HandlerFunc defines the request handler used by bee
/*
	提供给框架用户，用来定义路由映射的处理方法
	参数变为 Context
*/
type HandlerFunc func(ctx *Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	/*
		路由映射表
		其中 key 由请求方法和静态路由地址构成，例如 GET-/  GET-/hello  POST-/hello
		这样针对相同的路由，如果请求方法不同，可以映射不同的处理方法(Handler),value是用户映射的处理方法
	*/
	router *router
}

// New is the constructor of bee.Engine
func New() *Engine {
	return &Engine{router: newRouter()}
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

/*
	在 go 中，实现了接口方法的 struct 都可以强制转换为接口类型
	依然是实现了 ServeHTTP 接口，接管了所有的HTTP请求。
*/
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	/*
		解析请求的路径 查找路由映射表
		如果查到，就执行注册的处理方法，如果查不到，就返回 404 NOT FOUND
	*/
	c := newContext(w, req)
	engine.router.handle(c)
}
