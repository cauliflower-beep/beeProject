仿照极客兔兔，7天从0实现系列第一篇——web框架所写。

原文地址：

[7天用Go从零实现Web框架Gee教程 | 极客兔兔 (geektutu.com)](https://geektutu.com/post/gee.html)

## Day0 序言

### 设计一个框架

大部分时候，我们要开发一个 Web 应用，第一反应是选择使用哪个框架。

不同框架设计的理念、提供的功能有很大的差别，比如 Python 的 `django`和`flask`，前者大而全，后者小而美。Go语言/golang 也是如此，新框架层出不穷，`Beego`，`Gin`，`Iris`等。

为什么必须使用框架，不直接使用标准库呢？

回答设计框架的必要性之前，我们需要搞清楚一个核心问题：

> 框架为我们解决了什么痛点？

只有弄明白这一点，才能理清我们需要在框架中实现什么功能。

我们先看看标准库`net/http`如何处理一个请求。

```go
func main() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/count", counter)
    log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
}
```

`net/http`提供了基础的Web功能：**监听端口**，**映射静态路由**，**解析HTTP报文**。另外一些Web开发中常见的需求并不支持，需要手工实现，例如：

- **动态路由**：例如`hello/:name`，`hello/*`这类的规则。
- **鉴权**：没有分组/统一鉴权的能力，需要在每个路由映射的handler中实现。
- **模板**：没有统一简化的HTML机制。
- …

当我们离开框架，使用基础库时，需要频繁手工处理的地方，就是框架的价值所在；但并不是每一个频繁处理的地方都适合在框架中完成。

Python有一个很著名的Web框架，名叫[bottle](https://github.com/bottlepy/bottle)，整个框架由`bottle.py`一个文件构成，共4400行，可以说是一个微框架。理解这个微框架提供的特性，可以帮助我们理解框架的核心能力：

- **路由(Routing)**：将请求映射到函数，支持动态路由。例如`/hello/:name`。
- **模板(Templates)**：使用内置模板引擎提供模板渲染机制。
- **工具集(Utils)**：提供对 cookies，headers 等处理机制。
- **插件(Plugin)**：Bottle本身功能有限，但提供了插件机制。可以选择安装到全局，也可以只针对某几个路由生效。
- …

### Bee 框架

这个教程使用 Go 开发一个简单的 Web 框架，起名叫做`Bee`，在我学习Go语言的过程中，接触最多的 Web 框架是`Gin`，它的代码总共14K，其中测试代码9K，实际代码量只有**5K**。`Gin`与Python中的`Flask`很像，小而美。

**7天实现Bee框架**这个教程的很多设计，包括源码，参考了`Gin`，大家可以看到很多`Gin`的影子。

时间关系，同时为了尽可能地简洁明了，这个框架中的很多部分实现的功能都很简单，但是尽可能地体现一个框架核心的设计原则。例如`Router`的设计，虽然支持的动态路由规则有限，但为了性能考虑匹配算法是用`Trie树`实现的，`Router`最重要的指标之一便是性能。

希望这个教程能够对你有所启发。

## Day1 Http基础

原文地址：[前置知识(http.Handler接口)](https://geektutu.com/post/gee-day1.html)

本文内容：

- 简单介绍`net/http`库以及`http.Handler`接口。
- 搭建`Bee`框架的雏形，**代码约50行**。

### 标准库启动Web服务

Go语言内置了 `net/http`库，封装了HTTP网络编程的基础接口，我们实现的`Bee` Web 框架便是基于`net/http`的。首先通过一个例子，简单介绍下这个库的使用。

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

// handler echoes r.URL.Path
func indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
}

// handler echoes r.URL.Header
func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
}
```

我们设置了2个路由：`/`和`/hello`，分别绑定 *indexHandler* 和 *helloHandler* ， 根据不同的HTTP请求会调用不同的处理函数。访问`/`，响应是`URL.Path = /`，而`/hello`的响应则是请求头(header)中的键值对信息。

用 curl 工具测试一下，将会得到如下的结果。

```go
$ curl http://localhost:9999/
URL.Path = "/"
$ curl http://localhost:9999/hello
Header["Accept"] = ["*/*"]
Header["User-Agent"] = ["curl/7.54.0"]
```

*main* 函数的最后一行，是用来启动 Web 服务的。

首先是地址，`:9999`表示在 *9999* 端口监听；而第二个参数代表处理所有的HTTP请求的实例，`nil` 代表使用标准库中的实例处理。

重点在于第二个参数，它是我们基于`net/http`标准库实现Web框架的入口。

### 实现http.Handler接口

```go
package http

type Handler interface {
    ServeHTTP(w ResponseWriter, r *Request)
}

func ListenAndServe(address string, h Handler) error
```

上文说到，第二个参数代表处理Http请求的实例，那它的类型是什么呢？

通过查看`net/http`的源码可以发现，`Handler`是一个接口，需要实现方法 *ServeHTTP* ，也就是说，只要传入任何实现了 *ServerHTTP* 接口的实例，所有的HTTP请求，就都交给了该实例处理了。马上来试一试吧。

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

// Engine is the uni handler for all requests
type Engine struct{}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	case "/hello":
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	default:
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}

func main() {
	engine := new(Engine)
	log.Fatal(http.ListenAndServe(":9999", engine))
}
```

我们定义了一个空结构体`Engine`，实现了方法`ServeHTTP`。这个方法有2个参数，第二个参数是 *Request* ，该对象包含了该HTTP请求的所有的信息，例如**请求地址**、**Header**、**Body**等信息；第一个参数是 *ResponseWriter* ，包含一组方法的接口。利用 *ResponseWriter* 可以构造针对该请求的响应。

在 *main* 函数中，我们给 *ListenAndServe* 方法的第二个参数传入了刚才创建的`engine`实例。

至此，我们走出了实现Web框架的第一步，即，**将所有的HTTP请求转向了我们自己的处理逻辑**。还记得吗，在实现`Engine`之前，我们调用 *http.HandleFunc* 实现了路由和Handler的映射，也就是只能针对具体的路由写处理逻辑。比如`/hello`。但是在实现`Engine`之后，我们拦截了所有的HTTP请求，拥有了统一的控制入口。在这里我们可以自由定义路由映射的规则，也可以统一添加一些处理逻辑，例如日志、异常处理等。

代码的运行结果与之前的是一致的。

### Bee框架的雏形

我们接下来重新组织上面的代码，搭建出整个框架的雏形。

最终的代码目录结构是这样的。

```go
bee/
  |--bee.go
  |--go.mod
main.go
go.mod
```

> go.mod

```go
module example

go 1.13

require bee v0.0.0

replace bee => ./bee
```

在 `go.mod` 中使用 `replace` 将 bee 指向 `./bee`。从 go 1.11 版本开始，引用相对路径的 package 需要使用上述方式。

> main.go

```go
package main

import (
	"fmt"
	"net/http"

	"bee"
)

func main() {
	r := bee.New()
	r.GET("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	})

	r.GET("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})

	r.Run(":9999")
}
```

如果你使用过`gin`框架的话，肯定会觉得无比的亲切。

`bee`框架的设计以及API均参考了`gin`。使用`New()`创建 bee 的实例，使用 `GET()`方法添加路由，最后使用`Run()`启动Web服务。这里的路由，只是静态路由，不支持`/hello/:name`这样的动态路由，动态路由我们将在下一次实现。

> bee.go

```go
package bee

import (
	"fmt"
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

type Engine struct {
	router map[string]HandlerFunc
}

func New() *Engine {
	return &Engine{router: make(map[string]HandlerFunc)}
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	engine.router[key] = handler
}

func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.Method + "-" + req.URL.Path
	if handler, ok := engine.router[key]; ok {
		handler(w, req)
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}

```

`bee.go`是重头戏，我们重点介绍一下这部分的实现。

- 首先定义了类型`HandlerFunc`，这是提供给框架用户的，用来定义路由映射的处理方法。我们在`Engine`中，添加了一张路由映射表`router`，key 由请求方法和静态路由地址构成，例如`GET-/`、`GET-/hello`、`POST-/hello`，这样针对相同的路由，如果请求方法不同,可以映射不同的处理方法(Handler)，value 是用户映射的处理方法。
- 当用户调用`(*Engine).GET()`方法时，会将路由和处理方法注册到映射表 *router* 中，`(*Engine).Run()`方法，是 *ListenAndServe* 的包装。
- `Engine`实现的 *ServeHTTP* 方法的作用就是，解析请求的路径，查找路由映射表，如果查到，就执行注册的处理方法。如果查不到，就返回 *404 NOT FOUND* 。

执行`go run main.go`，再用 *curl* 工具访问，结果与最开始的一致。

```go
$ curl http://localhost:9999/
URL.Path = "/"
$ curl http://localhost:9999/hello
Header["Accept"] = ["*/*"]
Header["User-Agent"] = ["curl/7.54.0"]
curl http://localhost:9999/world
404 NOT FOUND: /world
```

至此，整个`Bee`框架的原型已经出来了。实现了**路由映射表**，提供了用户注册静态路由的方法，包装了启动服务的函数。

当然，到目前为止，我们还没有实现比`net/http`标准库更强大的能力，不用担心，很快就可以将动态路由、中间件等功能添加上去了。

### 自我总结

开发一个框架，俗称“造轮子”，都是基于原始的标准库扩展而来。所以首先我们要对标准库非常熟悉，并且标准库满足不了我们的常用需求。

这节课最大的设计是实现了engine结构体，接管所有的http请求并处理。

## Day2 上下文Context

原文地址：[上下文设计(Context)](https://geektutu.com/post/gee-day2.html)

本文内容：

- 将`路由(router)`独立出来，方便之后增强。
- 设计`上下文(Context)`，封装 Request 和 Response ，提供对 JSON、HTML 等返回类型的支持。
- **框架代码140行，新增代码约90行**

### 使用效果

为了展示第二天的成果，我们先看一看在使用时的效果。

```go
func main() {
	r := bee.New()
	r.GET("/", func(c *bee.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Bee</h1>")
	})
	r.GET("/hello", func(c *bee.Context) {
		// expect /hello?name=goku
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.POST("/login", func(c *bee.Context) {
		c.JSON(http.StatusOK, bee.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})

	r.Run(":9999")
}
```

首先，`Handler`的参数变成成了`bee.Context`，提供了查询Query/PostForm参数的功能。

其次，`bee.Context`封装了`HTML/String/JSON`函数，能够快速构造HTTP响应。

### 设计Context

#### 必要性

Web服务的**本质**，无非是根据请求`*http.Request`，构造响应`http.ResponseWriter`。

但是这两个对象提供的接口粒度太细，比如我们要构造一个完整的响应，需要考虑**消息头**(Header)和**消息体**(Body)，而 Header 包含了**状态码**(StatusCode)，**消息类型**(ContentType)等几乎每次请求都需要设置的信息。因此，如果不进行有效的封装，那么框架的用户将需要写大量重复，繁杂的代码，还容易出错。

针对常用场景，能够**高效地构造出 HTTP 响应**是一个好的框架必须考虑的点。用返回 JSON 数据作比较，感受下封装前后的差距：

> 封装前

```go
obj = map[string]interface{}{
    "name": "geektutu",
    "password": "1234",
}
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
encoder := json.NewEncoder(w)
if err := encoder.Encode(obj); err != nil {
    http.Error(w, err.Error(), 500)
}
```

> 封装后：

```go
c.JSON(http.StatusOK, bee.H{
    "username": c.PostForm("username"),
    "password": c.PostForm("password"),
})
```

针对使用场景，封装`*http.Request`和`http.ResponseWriter`的方法，简化相关接口的调用，只是设计 Context 的原因之一。

对于框架来说，还需要支撑**额外的功能**。例如，将来解析动态路由`/hello/:name`，参数`:name`的值放在哪呢？再比如，框架需要支持中间件，那中间件产生的信息放在哪呢？

Context **随着每一个请求的出现而产生，请求的结束而销毁**。和当前请求强相关的信息都应由 Context 承载。

因此，设计 Context 结构，扩展性和复杂性留在了内部，而对外简化了接口。路由的处理函数，以及将要实现的中间件，参数都统一使用 Context 实例， Context 就像一次会话的百宝箱，可以找到任何东西。

#### 具体实现

```go
type H map[string]interface{}

type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	// response info
	StatusCode int
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
```

- 代码最开头，给`map[string]interface{}`起了一个别名`bee.H`，构建JSON数据时，显得更简洁。
- `Context`目前只包含了`http.ResponseWriter`和`*http.Request`，另外提供了对 Method 和 Path 这两个常用属性的直接访问。
- 提供了访问Query和PostForm参数的方法。
- 提供了快速构造String/Data/JSON/HTML响应的方法。

### 路由(Router)

我们将和路由相关的方法和结构提取出来，放到了一个新的文件中`router.go`，方便我们下一次对 router 的功能进行增强，例如提供动态路由的支持。 router 的 handle 方法作一个细微的调整，即 handler 的参数，变成了 Context。

```go
type router struct {
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{handlers: make(map[string]HandlerFunc)}
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	log.Printf("Route %4s - %s", method, pattern)
	key := method + "-" + pattern
	r.handlers[key] = handler
}

func (r *router) handle(c *Context) {
	key := c.Method + "-" + c.Path
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
```

### 框架入口

```go
// HandlerFunc defines the request handler used by bee
type HandlerFunc func(*Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
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

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}
```

将`router`相关的代码独立后，`bee.go`简单了不少。最重要的还是通过实现 ServeHTTP 接口，接管了所有的 HTTP 请求。

相比第一天的代码，这个方法也有细微的调整，在调用 router.handle 之前，构造了一个 Context 对象。这个对象目前还非常简单，仅仅是包装了原来的两个参数，之后我们会慢慢地给Context插上翅膀。

如何使用，`main.go`一开始就已经亮相了。运行`go run main.go`，借助 curl ，一起看一看今天的成果吧。

```go
$ curl -i http://localhost:9999/
HTTP/1.1 200 OK
Date: Mon, 12 Aug 2019 16:52:52 GMT
Content-Length: 18
Content-Type: text/html; charset=utf-8
<h1>Hello bee</h1>

$ curl "http://localhost:9999/hello?name=beektutu"
hello beektutu, you're at /hello

$ curl "http://localhost:9999/login" -X POST -d 'username=beektutu&password=1234'
{"password":"1234","username":"beektutu"}

$ curl "http://localhost:9999/xxx"
404 NOT FOUND: /xxx
```

### 自我总结

本文探讨了封装context的必要性，一是可以简化重复代码，把复杂性留在内部，对外暴露简单易用的接口；二是承载某次会话过程中所有的信息，方便后续使用。

另外还抽离了路由模块，方便后续对路由功能进行扩展。

## Day3 前缀树路由

原文链接：[Trie树路由(Router)](https://geektutu.com/post/gee-day3.html)

本文内容：

- 使用 **Trie 树**实现动态路由(dynamic route)解析。
- 支持两种模式`:name`和`*filepath`，**代码约150行**。

### Trie 树简介

前面的章节中，我们用了`map`结构存储路由表，索引非常高效，但是弊端也很明显：**键值对的存储的方式，只能用来索引静态路由**。如果我们想支持类似于`/hello/:name`这样的动态路由怎么办？

> 所谓**动态路由**，即一条路由规则可以匹配**某一类型而非某一条固定的路由**。例如`/hello/:name`，可以匹配`/hello/goku`、`hello/tom`等。

动态路由有很多种实现方式，支持的规则、性能等有很大的差异。例如开源的路由实现`gorouter`支持在路由规则中嵌入正则表达式，像`/p/[0-9A-Za-z]+`，即路径中的参数仅匹配数字和字母；另一个开源实现`httprouter`就不支持正则表达式。

著名的Web开源框架`gin` 在早期的版本，并没有实现自己的路由，而是直接使用了`httprouter`，后来不知道什么原因，放弃了`httprouter`，自己实现了一个版本。

![trie tree](.\imgs\trie_eg.jpg)

实现动态路由最常用的数据结构，被称为**前缀树**(Trie树)：每一个节点的所有的子节点都拥有相同的前缀。这种结构非常适用于路由匹配，比如我们定义了如下路由规则：

- /:lang/doc
- /:lang/tutorial
- /:lang/intro
- /about
- /p/blog
- /p/related

我们用前缀树来表示，是这样的。

![trie tree](.\imgs\trie_router.png)

HTTP请求的路径是由`/`分隔的多段构成的，因此，每一段可以作为前缀树的一个节点。通过树结构查询，如果中间某一层的节点都不满足条件，那么就说明没有匹配到的路由，查询结束。

接下来我们实现的动态路由具备以下两个功能。

- 参数匹配`:`。例如 `/p/:lang/doc`，可以匹配 `/p/c/doc` 和 `/p/go/doc`。
- 通配`*`。例如 `/static/*filepath`，可以匹配`/static/fav.ico`，也可以匹配`/static/js/jQuery.js`，这种模式常用于静态服务器，能够递归地匹配子路径。

### Trie 树实现

首先我们需要设计树节点上应该存储那些信息。

```go
type node struct {
	pattern  string // 待匹配路由，例如 /p/:lang
	part     string // 路由中的一部分，例如 :lang
	children []*node // 子节点，例如 [doc, tutorial, intro]
	isWild   bool // 是否模糊匹配，part 含有 : 或 * 时为true，表示模糊匹配成功
}
```

为了实现动态路由匹配，我们加上了`isWild`这个参数。它有什么用呢？

以匹配 `/p/go/doc/`这个路由为例，第一层节点，`p`**精准**匹配到了`p`，第二层节点，`go`**模糊**匹配到`:lang`，那么将会把`lang`这个参数赋值为`go`，继续下一层匹配。

#### 辅助函数

我们将匹配的逻辑，包装为**辅助函数**：

```go
// 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}
// 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
```

对于路由来说，最重要的当然是**注册**与**匹配**了。开发服务时，注册路由规则，映射handler；访问时，匹配路由规则，查找到对应的handler。

因此，Trie 树需要支持节点的插入与查询，我们分别来实现这两部分的功能。

#### 节点插入

插入节点很简单，还以`/p/:lang/doc`为例，只有在第三层节点，即`doc`节点，`pattern`才会设置为`/p/:lang/doc`，`p`和`:lang`节点的`pattern`属性皆为空。

```go
// 这里的height指的是当前节点的高度，从根节点算起
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
        // 如果没有匹配到当前`part`的节点，则新建一个
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
    // 递归插入
	child.insert(pattern, parts, height+1)
}
```

#### 节点匹配

路由匹配同样也需要递归查询每一层节点。

经过上面的路由注册逻辑，只有**访问路径**的最后一层节点才会存在完整的pattern。故我们可以使用`n.pattern == ""`来判断路由规则是否匹配成功。

例如，`/p/python`虽能成功匹配到`:lang`，但`:lang`的`pattern`值为空，因此匹配失败；

**通配符**也一样，递归查询每一层的节点，匹配到了`*`，匹配失败，退出；或者匹配到了第`len(parts)`层节点，匹配成功。

```go
// 这里的height同样指的是当前节点的高度，从根节点算起
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}
```

### Router

Trie 树的插入与查找都实现了，接下来我们将 Trie 树应用到路由中去。

我们使用 roots 来存储每种请求方式的Trie 树根节点。使用 handlers 存储每种请求方式的 HandlerFunc 。

getRoute 函数中，还解析了`:`和`*`两种匹配符的参数，返回一个 map 。例如`/p/go/doc`匹配到`/p/:lang/doc`，解析结果为：`{lang: "go"}`，`/static/css/beektutu.css`匹配到`/static/*filepath`，解析结果为`{filepath: "css/beektutu.css"}`。

```go
type router struct {
    // roots key eg. 
    // roots['GET'] roots['POST']
	roots    map[string]*node
    // handlers key eg. 
    // handlers['GET-/p/:lang/doc']
    // handlers['POST-/p/book']
	handlers map[string]HandlerFunc
}
 
func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// 路径解析 返回parts列表
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
            // 只允许出现一次通配符 *
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)

	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}
```

### Context与handle的变化

动态路由的场景中，路径参数往往包含了一些重要的请求信息，其对应的 HandlerFunc 也需要根据这些参数进行不同的逻辑处理。故 Context 对象需要增加一个属性和方法，提供对路由参数的访问。

我们将解析后的参数存储到`Params`中，通过类似`c.Param("lang")`的方式获取到对应的值。

```go
// context.go
type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	Params map[string]string
	// response info
	StatusCode int
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// router.go
func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
```

`router.go`的变化比较小，比较重要的一点是，在调用匹配到的`handler`前，将解析出来的路由参数赋值给`c.Params`。这样就能够在`handler`中，通过`Context`对象访问到具体的值了。

### 单元测试

```go
func newTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/assets/*filepath", nil)
	return r
}

func TestParsePattern(t *testing.T) {
	ok := reflect.DeepEqual(parsePattern("/p/:name"), []string{"p", ":name"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*"), []string{"p", "*"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*name/*"), []string{"p", "*name"})
	if !ok {
		t.Fatal("test parsePattern failed")
	}
}

func TestGetRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/hello/beektutu")

	if n == nil {
		t.Fatal("nil shouldn't be returned")
	}

	if n.pattern != "/hello/:name" {
		t.Fatal("should match /hello/:name")
	}

	if ps["name"] != "beektutu" {
		t.Fatal("name should be equal to 'beektutu'")
	}

	fmt.Printf("matched path: %s, params['name']: %s\n", n.pattern, ps["name"])

}
```

### 使用Demo

看看框架使用的样例吧。

```go
func main() {
	r := bee.New()
	r.GET("/", func(c *bee.Context) {
		c.HTML(http.StatusOK, "<h1>Hello bee</h1>")
	})

	r.GET("/hello", func(c *bee.Context) {
		// expect /hello?name=beektutu
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.GET("/hello/:name", func(c *bee.Context) {
		// expect /hello/beektutu
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})

	r.GET("/assets/*filepath", func(c *bee.Context) {
		c.JSON(http.StatusOK, bee.H{"filepath": c.Param("filepath")})
	})

	r.Run(":9999")
}
```

使用`curl`工具，测试结果。

```go
$ curl "http://localhost:9999/hello/beektutu"
hello beektutu, you're at /hello/beektutu

$ curl "http://localhost:9999/assets/css/beektutu.css"
{"filepath":"css/beektutu.css"}
```

### 自我总结

框架开发要从使用角度出发。不能一开始盲目的上手开发路由模块，会没有方向。

而是应该想，我怎么使用这个框架开发web服务？我是不是需要框架给我提供增加路由的功能？另外增加路由之后，框架是不是可以自动帮我匹配路由？
从这个角度出发，就有了开发方向了。

## Day4 分组控制Group

原文链接：[分组控制(Group)](https://geektutu.com/post/gee-day4.html)

本文内容：实现路由分组控制(Route Group Control)

### 分组的意义

**分组控制**(Group Control)是 Web 框架应提供的基础功能之一。所谓分组，是指路由的分组。如果没有路由分组，我们需要针对每一个路由进行控制。但是真实的业务场景中，往往某一组路由需要相似的处理。例如：

- 以`/post`开头的路由匿名可访问。
- 以`/admin`开头的路由需要鉴权。
- 以`/api`开头的路由是 RESTful 接口，可以对接第三方平台，需要三方平台鉴权。

大部分情况下的路由分组，是以相同的前缀来区分的。因此，我们今天实现的分组控制也是**以前缀来区分**，并且**支持分组的嵌套**。例如`/post`是一个分组，`/post/a`和`/post/b`可以是该分组下的子分组。作用在`/post`分组上的中间件(middleware)，也都会作用在子分组，子分组还可以应用自己特有的中间件。

中间件可以给框架提供无限的扩展能力，应用在分组上，可以使得分组控制的收益更为明显，而不是共享相同的路由前缀这么简单。例如`/admin`的分组，可以应用鉴权中间件；`/`分组应用日志中间件，`/`是默认的最顶层的分组，也就意味着给所有的路由，即整个框架增加了记录日志的能力。

提供扩展能力支持中间件的内容，我们将在下一节当中介绍。

### 分组嵌套

一个 Group 对象需要具备哪些属性呢？

首先是**前缀(prefix)**，比如`/`，或者`/api`；其次要支持分组嵌套，那么需要知道**当前分组的父亲(parent)**是谁；当然了，按照我们一开始的分析，中间件是应用在分组上的，那还需要存储应用在该分组上的**中间件(middlewares)**；最后，还记得，我们之前调用函数`(*Engine).addRoute()`来映射所有的路由规则和 Handler 。如果Group对象需要直接映射路由规则的话，比如我们想在使用框架时，这么调用：

```go
r := bee.New()
v1 := r.Group("/v1")
v1.GET("/", func(c *bee.Context) {
	c.HTML(http.StatusOK, "<h1>Hello bee</h1>")
})
```

那么Group对象，还需要有访问`Router`的能力，为了方便，我们可以在Group中，保存一个指针，指向`Engine`，整个框架的所有资源都是由`Engine`统一协调的，那么就可以通过`Engine`间接地访问各种接口了。

所以，最后的 Group 的定义是这样的：

```go
RouterGroup struct {
	prefix      string // 前缀
	middlewares []HandlerFunc // 支持中间件
	parent      *RouterGroup  // 支持嵌套
	engine      *Engine       // 共享 Engine 实例，具备访问router的能力
}
```

我们还可以进一步地抽象，将`Engine`作为最顶层的分组，也就是说`Engine`拥有`RouterGroup`所有的能力。

```go
Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup // 存储所有分组
}
```

那我们就可以将和路由有关的函数，都交给`RouterGroup`实现了。

```go
// New is the constructor of bee.Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Group 创建一个新的路由分组
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine, // 所有的路由分组共享同一个engine实例
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
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
```

可以仔细观察下`addRoute`函数，调用了`group.engine.router.addRoute`来实现了路由的映射。由于`Engine`从某种意义上继承了`RouterGroup`的所有属性和方法，因为 (*Engine).engine 是指向自己的。这样实现，我们既可以像原来一样添加路由，也可以通过分组添加路由。

### 使用 Demo

测试框架的Demo就可以这样写了：

```go
func main() {
	r := bee.New()
	r.GET("/index", func(c *bee.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	v1 := r.Group("/v1")
	{
		v1.GET("/", func(c *bee.Context) {
			c.HTML(http.StatusOK, "<h1>Hello bee</h1>")
		})

		v1.GET("/hello", func(c *bee.Context) {
			// expect /hello?name=beektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *bee.Context) {
			// expect /hello/beektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *bee.Context) {
			c.JSON(http.StatusOK, bee.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})

	}

	r.Run(":9999")
}
```

通过 curl 简单测试：

```go
$ curl "http://localhost:9999/v1/hello?name=beektutu"
hello beektutu, you're at /v1/hello

$ curl "http://localhost:9999/v2/hello/beektutu"
hello beektutu, you're at /hello/beektutu
```

- 第五天：[中间件(Middleware)](https://geektutu.com/post/gee-day5.html)，[Code - Github](https://github.com/geektutu/7days-golang/tree/master/gee-web/day5-middleware)
- 第六天：[HTML模板(Template)](https://geektutu.com/post/gee-day6.html)，[Code - Github](https://github.com/geektutu/7days-golang/tree/master/gee-web/day6-template)
- 第七天：[错误恢复(Panic Recover)](https://geektutu.com/post/gee-day7.html)，[Code - Github](https://github.com/geektutu/7days-golang/tree/master/gee-web/day7-panic-recover)