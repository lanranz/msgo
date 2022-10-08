package msgo

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
)

//上下文，用于传递信息
type Context struct {
	W      http.ResponseWriter
	R      *http.Request
	engine *Engine
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

func (c *Context) Template(name string, data any) error {
	c.W.Header().Set("Content-Type", "text/html;charset=utf-8")
	err := c.engine.HTMLRender.Template.ExecuteTemplate(c.W, name, data)
	return err
}

func (c *Context) JSON(status int, data any) error {
	c.W.Header().Set("Content-Type", "text/json;charset=utf-8")
	c.W.WriteHeader(status)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.W.Write(jsonData)
	return err
}

func (c *Context) XML(status int, data any) error {
	c.W.Header().Set("Content-Type", "text/xml;charset=utf-8")
	c.W.WriteHeader(status)
	//xmlData, err := xml.Marshal(data)
	//if err != nil {
	//	return err
	//}
	//_, err = c.W.Write(xmlData)
	err := xml.NewEncoder(c.W).Encode(data)
	return err
}

//下载的文件名是请求的名字，不能自定义
func (c *Context) File(fileName string) {
	http.ServeFile(c.W, c.R, fileName)
}

//能够自定义返回文件的名称
//当参数都是string类型时可以少写一个string
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

//从文件系统获取，需要指定文件系统路径，filepath是文件系统的相对路径
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

//重定向
func (c *Context) Redirect(status int, location string) {
	//对状态码进行判断
	if (status < http.StatusMultipleChoices || status > http.StatusPermanentRedirect) && status != http.StatusCreated {
		panic(fmt.Sprintf("Cannot redirect with status code %d", status))
	}
	http.Redirect(c.W, c.R, location, status)
}

//支持通过占位符的方式输出string
func (c *Context) String(status int, format string, values ...any) (err error) {
	plainContentType := "text/plain; charset=utf-8"
	c.W.Header().Set("Content-Type", plainContentType)
	c.W.WriteHeader(status)
	//大于0说明有占位符，不是纯字符串
	if len(values) > 0 {
		_, err = fmt.Fprintf(c.W, format, values...)
		return err
	}
	_, err = c.W.Write(StringToBytes(format))
	return
}
