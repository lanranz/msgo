package msgo

import (
	"html/template"
	"net/http"
)

//上下文，用于传递信息
type Context struct {
	W http.ResponseWriter
	R *http.Request
}

//返回的页面
func (c *Context) HTML(status int, html string) error {
	c.W.Header().Set("Content-Type", "text/html;charset=utf-8")
	//设置状态是200.默认不设置的话，如果调用了write这个方法，实际上默认返回状态200,表示响应请求成功。
	c.W.WriteHeader(status)
	_, err := c.W.Write([]byte(html))
	return err
}

//支持模板
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

//glob加载所有模板，不需要一个个写需要的模板
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
