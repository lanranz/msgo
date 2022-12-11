package msgo

import "net/http"

func Recovery(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx.Logger.Error(err)
				ctx.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		next(ctx)
	}
}
