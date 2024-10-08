package bee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// H 取别名，构建JSON 数据时，显得更简洁
type H map[string]interface{}

type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info 提供对 Path Method 这两个常用属性的直接访问
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

/*-------完整响应需要考虑的基本信息-----------*/

func (c *Context) Status(code int) {
	c.StatusCode = code
	// 设置http回包数据中的响应行
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	// 设置http回包数据中的响应头
	c.Writer.Header().Set(key, value)
}

/*------快速构造响应-------*/
func (c *Context) String(code int, format string, values ...interface{}) {
	// 注意调用顺序，必须是 Header().Set -> WriteHeader -> Write
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	// 设置http回包数据中的响应体
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	// 创建一个json编码器，将数据编码并写入c.Writer
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Date(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
