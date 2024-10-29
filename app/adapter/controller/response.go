package controller

import (
	"clean-serverless-book-sample/logger"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// commonHeaders 各レスポンスに共通で含むヘッダー
func commonHeaders(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")
	ctx.Header("Access-Control-Allow-Origin", "*")
}

// Response200 JSONを含めた200レスポンス
func Response200(ctx *gin.Context, body interface{}) {
	commonHeaders(ctx)
	ctx.JSON(http.StatusOK, body)
}

// Response200OK okメッセージを含めた200レスポンス
func Response200OK(ctx *gin.Context) {
	commonHeaders(ctx)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}

// Response201 IDを含めた201レスポンス
func Response201(ctx *gin.Context, id uint64) {
	commonHeaders(ctx)
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "OK",
		"id":      fmt.Sprintf("%d", id),
	})
}

// Response400 エラーメッセージを含めた400レスポンス
func Response400(ctx *gin.Context, errs map[string]error) {
	log := logger.GetLogger()
	log.Warn("Validation errors occurred", "errors", errs)
	commonHeaders(ctx)
	ctx.JSON(http.StatusBadRequest, gin.H{
		"message": "入力値を確認してください。",
		"errors":  ConvertErrorsToMessage(errs),
	})
}

// Response404 404レスポンス
func Response404(ctx *gin.Context) {
	commonHeaders(ctx)
	ctx.JSON(http.StatusNotFound, gin.H{
		"message": "結果が見つかりません。",
	})
}

// Response500 500レスポンス
func Response500(ctx *gin.Context, err error) {
	log := logger.GetLogger()
	log.Error("Internal server error", "error", err)
	commonHeaders(ctx)
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"message": "サーバエラーが発生しました。",
	})
}
