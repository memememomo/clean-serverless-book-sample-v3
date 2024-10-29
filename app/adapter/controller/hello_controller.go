package controller

import (
	"clean-serverless-book-sample/registry"
	"clean-serverless-book-sample/usecase"
	"encoding/json"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type HelloController struct {
	log *slog.Logger
}

// PostHelloRequest HTTPリクエストのJSON形式を表した構造体
type PostHelloRequest struct {
	Name string `json:"name"`
}

// HelloMessageResponse HTTPレスポンスのJSON形式を表した構造体
type HelloMessageResponse struct {
	Message string `json:"message"`
}

// ValidateHelloMessageSettings バリデーションの設定
func ValidateHelloMessageSettings() *Validator {
	return &Validator{
		Settings: []*ValidatorSetting{
			{ArgName: "name", ValidateTags: "required"},
		},
	}
}

// PostHello コントローラの実装
func (ctrl *HelloController) PostHello(ctx *gin.Context) {
	// リクエストボディを取得
	body, err := ctx.GetRawData()
	if err != nil {
		Response500(ctx, err)
		return
	}

	// バリデーション
	validator := ValidateHelloMessageSettings()
	validErr := validator.ValidateBody(string(body))
	if validErr != nil {
		ctrl.log.Warn("Validation error", "error", validErr)
		Response400(ctx, validErr)
		return
	}

	// JSONから構造体に変換
	var req PostHelloRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		Response500(ctx, err)
		return
	}

	// UseCaseを実行
	h := registry.GetFactory().BuildCreateHelloMessage()
	res, err := h.Execute(&usecase.CreateHelloMessageRequest{
		Name: req.Name,
	})
	if err != nil {
		Response500(ctx, err)
		return
	}

	// HTTPレスポンスを返す
	Response200(ctx, &HelloMessageResponse{
		Message: res.Message,
	})
}
