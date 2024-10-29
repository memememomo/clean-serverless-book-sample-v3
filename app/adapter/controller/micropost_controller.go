package controller

import (
	"clean-serverless-book-sample/domain"
	"clean-serverless-book-sample/registry"
	"clean-serverless-book-sample/usecase"
	"clean-serverless-book-sample/utils"
	"encoding/json"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type MicropostController struct {
	log *slog.Logger
}

// MicropostSettingsValidator バリデーション設定
func MicropostSettingsValidator() *Validator {
	return &Validator{
		Settings: []*ValidatorSetting{{ArgName: "content", ValidateTags: "required,max=140"}},
	}
}

// RequestMicropost HTTPリクエストで送られてくるJSON形式を表した構造体
type RequestMicropost struct {
	Content string `json:"content"`
}

// RequestPostMicropost PostMicropostのリクエスト
type RequestPostMicropost struct {
	RequestMicropost
}

// RequestPutMicropost PutMicropostのリクエスト
type RequestPutMicropost struct {
	RequestMicropost
}

// ResponseMicropost レスポンス用のJSON形式を表した構造体
type ResponseMicropost struct {
	ID      uint64 `json:"id"`
	UserID  uint64 `json:"user_id"`
	Content string `json:"content"`
}

// ResponseMicroposts Micropostリストレスポンス用のJSON形式を表した構造体
type ResponseMicroposts struct {
	Microposts []*ResponseMicropost `json:"microposts"`
}

// PostMicroposts 新規作成
func (ctrl *MicropostController) PostMicroposts(ctx *gin.Context) {
	ctrl.log.Info("Starting PostMicroposts handler")

	// リクエストボディを取得
	body, err := ctx.GetRawData()
	if err != nil {
		ctrl.log.Error("Failed to get request body", "error", err)
		Response500(ctx, err)
		return
	}

	// バリデーション処理
	validator := MicropostSettingsValidator()
	validErr := validator.ValidateBody(string(body))
	if validErr != nil {
		ctrl.log.Warn("Validation error", "error", validErr)
		Response400(ctx, validErr)
		return
	}

	// パスパラメータからユーザーIDを取得する
	userID, err := utils.ParseUint(ctx.Param("user_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse user_id", "error", err)
		Response500(ctx, err)
		return
	}

	// JSON形式から構造体に変換
	var req RequestPostMicropost
	err = json.Unmarshal(body, &req)
	if err != nil {
		ctrl.log.Error("Failed to unmarshal request body", "error", err)
		Response500(ctx, err)
		return
	}

	// 新規作成処理
	ctrl.log.Info("Creating new micropost", "userID", userID, "content", req.Content)
	creator := registry.GetFactory().BuildCreateMicropost()
	res, err := creator.Execute(&usecase.CreateMicropostRequest{
		Content: req.Content,
		UserID:  userID,
	})
	if err != nil {
		ctrl.log.Error("Failed to create micropost", "error", err)
		Response500(ctx, err)
		return
	}

	ctrl.log.Info("Successfully created micropost", "micropostID", res.MicropostID)
	// 201レスポンス
	Response201(ctx, res.MicropostID)
}

// PutMicropost 更新
func (ctrl *MicropostController) PutMicropost(ctx *gin.Context) {
	ctrl.log.Info("Starting PutMicropost handler")

	// リクエストボディを取得
	body, err := ctx.GetRawData()
	if err != nil {
		ctrl.log.Error("Failed to get request body", "error", err)
		Response500(ctx, err)
		return
	}

	// バリデーション処理
	validator := MicropostSettingsValidator()
	validErr := validator.ValidateBody(string(body))
	if validErr != nil {
		ctrl.log.Warn("Validation error", "error", validErr)
		Response400(ctx, validErr)
		return
	}

	// パスパラメータからユーザーIDを取得する
	userID, err := utils.ParseUint(ctx.Param("user_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse user_id", "error", err)
		Response500(ctx, err)
		return
	}

	// パスパラメータからマイクロポストIDを取得する
	micropostID, err := utils.ParseUint(ctx.Param("micropost_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse micropost_id", "error", err)
		Response500(ctx, err)
		return
	}

	// JSON形式から構造体に変換
	var req RequestPutMicropost
	err = json.Unmarshal(body, &req)
	if err != nil {
		ctrl.log.Error("Failed to unmarshal request body", "error", err)
		Response500(ctx, err)
		return
	}

	// 更新処理
	ctrl.log.Info("Updating micropost", "micropostID", micropostID, "userID", userID)
	updater := registry.GetFactory().BuildUpdateMicropost()
	_, err = updater.Execute(&usecase.UpdateMicropostRequest{
		Content:     req.Content,
		UserID:      userID,
		MicropostID: micropostID,
	})
	if err != nil {
		ctrl.log.Error("Failed to update micropost", "error", err)
		Response500(ctx, err)
		return
	}

	ctrl.log.Info("Successfully updated micropost", "micropostID", micropostID)
	// 200レスポンス
	Response200OK(ctx)
}

// GetMicroposts 一覧取得
func (ctrl *MicropostController) GetMicroposts(ctx *gin.Context) {
	ctrl.log.Info("Starting GetMicroposts handler")

	// パスパラメータからユーザーIDを取得
	userID, err := utils.ParseUint(ctx.Param("user_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse user_id", "error", err)
		Response500(ctx, err)
		return
	}

	// マイクロポスト取得処理
	ctrl.log.Info("Getting micropost list", "userID", userID)
	getter := registry.GetFactory().BuildGetMicropostList()
	res, err := getter.Execute(&usecase.GetMicropostListRequest{
		UserID: userID,
	})
	if err != nil {
		ctrl.log.Error("Failed to get micropost list", "error", err)
		Response500(ctx, err)
		return
	}

	// ドメインモデルからレスポンス用の構造体に詰め替える
	var resMicroposts = make([]*ResponseMicropost, len(res.Microposts))
	for i, m := range res.Microposts {
		resMicroposts[i] = &ResponseMicropost{
			ID:      m.ID,
			UserID:  m.UserID,
			Content: m.Content,
		}
	}

	ctrl.log.Info("Successfully retrieved micropost list", "count", len(resMicroposts))
	// レスポンス処理
	Response200(ctx, &ResponseMicroposts{
		Microposts: resMicroposts,
	})
}

// GetMicropost IDから取得
func (ctrl *MicropostController) GetMicropost(ctx *gin.Context) {
	ctrl.log.Info("Starting GetMicropost handler")

	// パスパラメータからユーザーIDを取得する
	userID, err := utils.ParseUint(ctx.Param("user_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse user_id", "error", err)
		Response500(ctx, err)
		return
	}

	// パスパラメータからマイクロポストIDを取得する
	micropostID, err := utils.ParseUint(ctx.Param("micropost_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse micropost_id", "error", err)
		Response500(ctx, err)
		return
	}

	// マイクロポスト取得処理
	ctrl.log.Info("Getting micropost by ID", "micropostID", micropostID, "userID", userID)
	getter := registry.GetFactory().BuildGetMicropostByID()
	res, err := getter.Execute(&usecase.GetMicropostByIDRequest{
		MicropostID: micropostID,
		UserID:      userID,
	})
	if err != nil {
		if err.Error() == domain.ErrNotFound.Error() {
			ctrl.log.Warn("Micropost not found", "micropostID", micropostID)
			Response404(ctx)
			return
		}
		ctrl.log.Error("Failed to get micropost", "error", err)
		Response500(ctx, err)
		return
	}

	ctrl.log.Info("Successfully retrieved micropost", "micropostID", res.Micropost.ID)
	// ドメインモデルからレスポンス用構造体に詰め替えて、レスポンス
	Response200(ctx, &ResponseMicropost{
		ID:      res.Micropost.ID,
		Content: res.Micropost.Content,
		UserID:  res.Micropost.UserID,
	})
}

// DeleteMicropost 削除処理
func (ctrl *MicropostController) DeleteMicropost(ctx *gin.Context) {
	ctrl.log.Info("Starting DeleteMicropost handler")

	// パスパラメータからユーザーIDを取得する
	userID, err := utils.ParseUint(ctx.Param("user_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse user_id", "error", err)
		Response500(ctx, err)
		return
	}

	// パスパラメータからマイクロポストIDを取得する
	micropostID, err := utils.ParseUint(ctx.Param("micropost_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse micropost_id", "error", err)
		Response500(ctx, err)
		return
	}

	// 削除処理
	ctrl.log.Info("Deleting micropost", "micropostID", micropostID, "userID", userID)
	deleter := registry.GetFactory().BuildDeleteMicropost()
	_, err = deleter.Execute(&usecase.DeleteMicropostRequest{
		MicropostID: micropostID,
		UserID:      userID,
	})
	if err != nil {
		ctrl.log.Error("Failed to delete micropost", "error", err)
		Response500(ctx, err)
	}

	ctrl.log.Info("Successfully deleted micropost", "micropostID", micropostID)
	// レスポンス
	Response200OK(ctx)
}
