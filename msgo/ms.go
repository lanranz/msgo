package msgo

import (
	"fmt"
	"html/template"
	"log"
	"msgo/render"
	"net/http"
	"sync"
)

const ANY = "ANY"

// 用type定义一个函数，函数的类型是func
type HandleFunc func(ctx *Context)

// 传入handlefunc并返回handlefunc，达到到在中间处理请求的目的
type MiddlewareFunc func(handleFunc HandleFunc) HandleFunc

// 路由组
type routerGroup struct {
	name string
	//map名 map [键类型]值类型，每个value都代表一个处理方法,这是一个多维映射，也就是map的嵌套，map的value可以为任意类型，
	//而key不行，其中第一个key为路由，映射一个value的key是请求方式，在映射一个处理方法
	handleFuncMap map[string]map[string]HandleFunc
	//只针对某个请求的中间件
	middlewaresFuncMap map[string]map[string][]MiddlewareFunc
	//请求方式，key:请求方式，value：路由
	handlerMethodMap map[string][]string
	treeNode         *treeNode
	//所有请求都执行的中间件
	Middlewares []MiddlewareFunc
}

type router struct {
	routerGroups []*routerGroup
}

// 创建路由组，输入值路由组名，返回路由组的map
func (r *router) Group(name string) *routerGroup {
	routerGroup := &routerGroup{
		name:               name,
		handleFuncMap:      make(map[string]map[string]HandleFunc),
		middlewaresFuncMap: make(map[string]map[string][]MiddlewareFunc),
		//key是请求名，值是路由组
		handlerMethodMap: make(map[string][]string),
		treeNode:         &treeNode{name: "/", children: make([]*treeNode, 0)},
	}
	//append(数组，需要添加的值)，将新建的路由组添加到router中的routerGroups数组中区
	r.routerGroups = append(r.routerGroups, routerGroup)
	return routerGroup
}

// ...代表可能有多个
func (r *routerGroup) Use(middlewareFunc ...MiddlewareFunc) {
	r.Middlewares = append(r.Middlewares, middlewareFunc...)
}

func (r *routerGroup) methodHandle(name string, method string, h HandleFunc, ctx *Context) {
	//组通用级别中间件
	if r.Middlewares != nil {
		for _, middlewareFunc := range r.Middlewares {
			h = middlewareFunc(h)
		}
	}
	//组路由级别的中间件
	middlewareFuncs := r.middlewaresFuncMap[name][method]
	if middlewareFuncs != nil {
		for _, middlewareFunc := range middlewareFuncs {
			h = middlewareFunc(h)
		}
	}
	h(ctx)
}

func (r *routerGroup) handle(name string, method string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	//判断map里面是否存在路由的name
	_, ok := r.handleFuncMap[name]
	if !ok {
		r.handleFuncMap[name] = make(map[string]HandleFunc)
		r.middlewaresFuncMap[name] = make(map[string][]MiddlewareFunc)
	}
	_, ok = r.handleFuncMap[name][method]
	if ok {
		panic("有重复的路由")
	}
	r.handleFuncMap[name][method] = handleFunc
	r.middlewaresFuncMap[name][method] = append(r.middlewaresFuncMap[name][method], middlewareFunc...)
	r.treeNode.Put(name)
}

// 创建路由，参数是路径和处理方法
func (r *routerGroup) Any(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, ANY, handleFunc, middlewareFunc...)
}

func (r *routerGroup) Get(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodGet, handleFunc, middlewareFunc...)
}

func (r *routerGroup) Post(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPost, handleFunc, middlewareFunc...)
}

func (r *routerGroup) Delete(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodDelete, handleFunc, middlewareFunc...)
}

func (r *routerGroup) Put(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPut, handleFunc, middlewareFunc...)
}

func (r *routerGroup) Patch(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPatch, handleFunc, middlewareFunc...)
}

func (r *routerGroup) Options(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodOptions, handleFunc, middlewareFunc...)
}

func (r *routerGroup) Head(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodHead, handleFunc, middlewareFunc...)
}

type Engine struct {
	//继承router结构体，Engine是router的子类
	router
	//提前加载模板到内存中，而不需要等到访问时才加载模板
	funcMap    template.FuncMap
	HTMLRender render.HTMLRender
	//sync.pool用于处理那些未来要用，但是暂时没有使用的值，这样可以不用重复分配内存，提高效率，这边用来解决context频繁每次调用都频繁创建的问题
	pool sync.Pool
}

func (e *Engine) setFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.setHtmlTemplate(t)
}

func (e *Engine) setHtmlTemplate(t *template.Template) {
	e.HTMLRender = render.HTMLRender{Template: t}
}

func New() *Engine {
	//初始化engine
	engine := &Engine{
		//初始化engine的父类路由，结构体加花括号的初始化是结构体初始化的常用写法
		router: router{},
	}
	//初始化一个pool,其中存放context
	engine.pool.New = func() any {
		return engine.allcateContext()
	}
	return engine
}

// 初始化sync.Pool中的context的内容
func (e *Engine) allcateContext() any {
	return &Context{engine: e}
}

// 实现ServerHTTP方法就相当于实现了源码中Handler这个接口，可以指定请求方式
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := e.pool.Get().(*Context)
	ctx.W = w
	ctx.R = r
	e.httpRequestHandle(ctx, w, r)
	e.pool.Put(ctx)
}

func (e *Engine) httpRequestHandle(ctx *Context, w http.ResponseWriter, r *http.Request) {
	method := r.Method

	for _, group := range e.routerGroups {
		routerName := SubStringLast(r.URL.Path, "/"+group.name)
		node := group.treeNode.Get(routerName)
		if node != nil && node.isEnd {
			handle, ok := group.handleFuncMap[node.routerName][ANY]
			if ok {
				group.methodHandle(node.routerName, ANY, handle, ctx)
				return
			}
			handle, ok = group.handleFuncMap[node.routerName][method]
			if ok {
				group.methodHandle(node.routerName, method, handle, ctx)
				return
			}
			//没有匹配到请求方法返回405状态
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "%s %s not allowed \n", r.RequestURI, method)
			return
		}
		//for name, methodHandle := range group.handleFuncMap {
		//	url := "/" + group.name + name
		//	//判断路由url是否匹配
		//	if r.RequestURI == url {
		//
		//	}
		//}
	}
	//放在循环之外，代表循环外的其他所有情况，404
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s %s not found \n", r.RequestURI, method)
}

func (e *Engine) Run() {
	//key:get value:func
	//上面将路由添加到map里面再通过遍历map处理路由
	//for _, group := range e.routerGroups {
	//	for key, value := range group.handleFuncMap {
	//		http.HandleFunc("/"+group.name+key, value)
	//	}
	//}
	//因为上边实现了engine类的ServeHandler方法，所以engine继承了Handel接口，可以写入参数中，所有的请求优先进入handle判断
	http.Handle("/", e)
	//启动8111端口服务器，ListenAndServe()函数有两个参数，当前监听的端口号和事件处理器Handler
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
