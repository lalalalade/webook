package pkg

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func WrapBody[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}
		// 下半段业务逻辑
		res, err := fn(ctx, req)
		if err != nil {

		}
		ctx.JSON(http.StatusOK, res)
	}
}

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
