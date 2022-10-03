package main

import (
	"fmt"
	"log"
	"msgo"
	"net/http"
)

type User struct {
	Name string
}

func Log(next msgo.HandleFunc) msgo.HandleFunc {
	return func(ctx *msgo.Context) {
		fmt.Println("打印请求参数 ")
		next(ctx)
		fmt.Println("返回执行时间")
	}
}

func main() {
	//http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
	//	fmt.Fprintln(writer, "hello mszlu.com")
	//})
	engine := msgo.New()
	g := engine.Group("user")
	g.Use(func(next msgo.HandleFunc) msgo.HandleFunc {
		return func(ctx *msgo.Context) {
			fmt.Println("pre handler")
			next(ctx)
			fmt.Println("post handler")
		}
	})
	g.Get("/hello", func(context *msgo.Context) {
		fmt.Println("handler")
		fmt.Fprintln(context.W, "get hello mszlu.com")
		//fmt.Fprintf(w, "%s 欢迎来到码神之路goweb教程", "mszlu.com")
	}, Log)
	g.Post("/hello", func(context *msgo.Context) {
		fmt.Fprintln(context.W, "post hello mszlu.com")
		//fmt.Fprintf(w, "%s 欢迎来到码神之路goweb教程", "mszlu.com")
	})
	g.Post("/info", func(context *msgo.Context) {
		fmt.Fprintln(context.W, "hello mszlu.com")
		//fmt.Fprintf(w, "%s 欢迎来到码神之路goweb教程", "mszlu.com")
	})
	g.Any("/any", func(context *msgo.Context) {
		fmt.Fprintln(context.W, "hello mszlu.com")
		//fmt.Fprintf(w, "%s 欢迎来到码神之路goweb教程", "mszlu.com")
	})
	g.Get("/get/:id", func(context *msgo.Context) {
		fmt.Fprintf(context.W, "%s get user info path variable", "hello mszlu.com")
		//fmt.Fprintf(w, "%s 欢迎来到码神之路goweb教程", "mszlu.com")
	})

	g.Get("/html", func(ctx *msgo.Context) {
		ctx.HTML(http.StatusOK, "<h1>码神之路</h1>")
	})

	g.Get("/htmlTemplate", func(ctx *msgo.Context) {
		user := &User{
			Name: "码神之路",
		}
		err := ctx.HTMLTemplate("login.html", user, "tpl/login.html", "tpl/header.html")
		if err != nil {
			log.Println(err)
		}
	})

	g.Get("/htmlTemplateGlob", func(ctx *msgo.Context) {
		user := &User{
			Name: "码神之路",
		}
		err := ctx.HTMLTemplateGlob("login.html", user, "tpl/*.html")
		if err != nil {
			log.Println(err)
		}
	})
	engine.Run()
}
