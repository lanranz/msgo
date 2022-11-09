package msgo

import (
	"html/template"
	"msgo/render"
	"net/http"
	"net/url"
)

// 上下文，用于传递信息
type Context struct {
	//http响应报文
	W http.ResponseWriter
	//请求报文
	R      *http.Request
	engine *Engine
	//存储url中的参数
	queryCache url.Values
}

// 获取url参数
func (c *Context) GetQueryCache(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

// 有多个url参数时获取
func (c *Context) GetQueryCacheArray(key string) ([]string, bool) {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	return values, ok
}

// 初始化url参数
func (c *Context) initQueryCache() {
	if c.queryCache == nil {
		if c.W != nil {
			c.queryCache = c.R.URL.Query()
		} else {
			c.queryCache = url.Values{}
		}
	}

}

// 返回的页面
func (c *Context) HTML(status int, html string) error {

	return c.Render(status, &render.HTML{Data: html, IsTemplate: false})
}

// 支持模板
func (c *Context) HTMLTemplate(name string, data any, filenames ...string) error {
	c.W.Header().Set("Content-Type", "text/html;charset=utf-8")
	//设置状态是200.默认不设置的话，如果调用了write这个方法，实际上默认返回状态200,表示响应请求成功。
	//new一个模板
	t := template.New(name)
	t, err := t.ParseFiles(filenames...)
	if err != nil {
		return err
	}
	err = t.Execute(c.W, data)
	return err
}

// glob加载所有模板，不需要一个个写需要的模板
func (c *Context) HTMLTemplateGlob(name string, data any, pattern string) error {
	c.W.Header().Set("Content-Type", "text/html;charset=utf-8")
	//设置状态是200.默认不设置的话，如果调用了write这个方法，实际上默认返回状态200,表示响应请求成功。
	//new一个模板
	t := template.New(name)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	err = t.Execute(c.W, data)
	return err
}

func (c *Context) Template(name string, data any) error {
	return c.Render(http.StatusOK, &render.HTML{
		Data:       data,
		IsTemplate: true,
		Template:   c.engine.HTMLRender.Template,
		Name:       name,
	})
}

func (c *Context) JSON(status int, data any) error {
	return c.Render(status, &render.JSON{Data: data})
}

func (c *Context) XML(status int, data any) error {
	return c.Render(status, &render.XML{
		Data: data,
	})
}

// 下载的文件名是请求的名字，不能自定义
func (c *Context) File(fileName string) {
	http.ServeFile(c.W, c.R, fileName)
}

// 能够自定义返回文件的名称
// 当参数都是string类型时可以少写一个string
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

// 从文件系统获取，需要指定文件系统路径，filepath是文件系统的相对路径
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

// 重定向
func (c *Context) Redirect(status int, url string) error {
	return c.Render(status, &render.Redirect{Code: status,
		Request:  c.R,
		Location: url,
	})
}

// 支持通过占位符的方式输出string
func (c *Context) String(status int, format string, values ...any) error {
	err := c.Render(status, &render.String{Format: format, Data: values})
	return err
}

func (c *Context) Render(statusCode int, r render.Render) error {
	err := r.Render(c.W)
	c.W.WriteHeader(statusCode)
	return err
}
