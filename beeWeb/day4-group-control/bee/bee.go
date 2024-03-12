package bee

import (
	"log"
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
		将 Engine 作为最顶层的分组 使得 Engine 拥有 RouterGroup 所有的能力
		这样就可以将和路由有关的函数，都交给 RouterGroup 实现了
	*/
	*RouterGroup
	/*
		路由映射表
		其中 key 由请求方法和静态路由地址构成，例如 GET-/  GET-/hello  POST-/hello
		这样针对相同的路由，如果请求方法不同，可以映射不同的处理方法(Handler),value是用户映射的处理方法
	*/
	router *router
	groups []*RouterGroup // store all groups
}

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // support middleware
	parent      *RouterGroup  // support nesting
	engine      *Engine       // all groups share a Engine instance 以获取访问 router 的能力
}

// New is the constructor of bee.Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Group is defined to create a new RouterGroup
// remember all groups share the same Engine instance
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	/*
		由于Engine 从某种意义上继承了 RouterGroup 的所有属性和方法，
		因为 (*Engine).engine 是指向自己的，这样实现，我们既可以像原来一样添加路由，也可以通过分组添加路由
	*/
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
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
