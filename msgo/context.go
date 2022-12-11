package msgo

import (
	"errors"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"msgo/binding"
	msLog "msgo/log"
	"msgo/render"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultMultipartMemory = 32 << 20 //默认最大内存32兆，<<是二进制左移20位

// 上下文，用于传递信息
type Context struct {
	//http响应报文
	W http.ResponseWriter
	//请求报文
	R      *http.Request
	engine *Engine
	//存储url中的参数
	queryCache url.Values
	//post表单参数
	formCache url.Values
	//json传参的传入json有多余字段校验功能的开关
	DisallowUnknownFields bool
	//json传参缺少字段检验开关
	IsValidate bool
	StatusCode int
	Logger     *msLog.Logger
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
	if c.W != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
}

func (c *Context) QueryArray(key string) (values []string) {
	c.initQueryCache()
	values, _ = c.queryCache[key]
	return
}

// 设置一个默认的参数值，当没有获取到参数时使用默认的值
func (c *Context) DefaultQuery(key, defaultValue string) string {
	array, ok := c.GetQueryCacheArray(key)
	if !ok {
		return defaultValue
	}
	return array[0]
}

func (c *Context) QueryMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetQueryMap(key)
	return
}

// url参数是map格式，类似于`http://localhost:8080/queryMap?user[id]=1&user[name]=张三`
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	dicts := make(map[string]string)
	exist := false
	for k, value := range m {
		//判断写法是否合规
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = value[0]
				//dicts[k[i+1:j]] = value[0]
			}
		}
	}
	return dicts, exist
}

func (c *Context) initPostFormcache() {
	if c.W != nil {
		if err := c.R.ParseMultipartForm(defaultMultipartMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
			}
		}
		c.formCache = c.R.PostForm
	} else {
		c.formCache = url.Values{}
	}
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) PostFormArray(key string) (values []string) {
	values, _ = c.GetPostFormArray(key)
	return
}

func (c *Context) PostFormMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetPostFormMap(key)
	return
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.initPostFormcache()
	values, ok := c.formCache[key]
	return values, ok
}

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initPostFormcache()
	return c.get(c.formCache, key)
}

func (c Context) FormFile(name string) *multipart.FileHeader {
	file, header, err := c.R.FormFile(name)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	return header
}

func (c Context) FormFiles(name string) []*multipart.FileHeader {
	multipartForm, err := c.MultipartForm()
	if err != nil {
		return make([]*multipart.FileHeader, 0)
	}
	return multipartForm.File[name]
}

func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(defaultMultipartMemory)
	return c.R.MultipartForm, err
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
	c.W.WriteHeader(statusCode)
	err := r.Render(c.W)
	c.StatusCode = statusCode
	return err
}

func (c *Context) BindJson(obj any) error {
	jsonBinding := binding.JSON
	jsonBinding.DisallowUnknownFields = c.DisallowUnknownFields
	jsonBinding.IsValidate = c.IsValidate
	return c.MustBindWith(obj, jsonBinding)
}

func (c *Context) BindXml(obj any) error {
	return c.MustBindWith(obj, binding.XML)
}

func (c *Context) MustBindWith(obj any, b binding.Binding) error {
	//如果发生错误，返回400状态码 参数错误
	if err := c.ShouldBindWith(obj, b); err != nil {
		c.W.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}

func (c *Context) ShouldBindWith(obj any, b binding.Binding) error {
	return b.Bind(c.R, obj)
}

func (c *Context) Fail(code int, msg string) {
	c.String(code, msg)
}
