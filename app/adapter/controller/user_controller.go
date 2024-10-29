package controller

import (
	"clean-serverless-book-sample/domain"
	"clean-serverless-book-sample/interactor"
	"clean-serverless-book-sample/registry"
	"clean-serverless-book-sample/usecase"
	"clean-serverless-book-sample/utils"
	"encoding/json"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type UserController struct {
	log *slog.Logger
}

// PostSettingValidator バリデーション設定
func PostSettingValidator() *Validator {
	return &Validator{
		Settings: []*ValidatorSetting{
			{ArgName: "user_name", ValidateTags: "required"},
			{ArgName: "email", ValidateTags: "required,email"},
		},
	}
}

// RequestPostUser PostUserのリクエスト
type RequestPostUser struct {
	Name  string `json:"user_name"`
	Email string `json:"email"`
}

// RequestPutUser PutUserのリクエスト
type RequestPutUser struct {
	Name  string `json:"user_name"`
	Email string `json:"email"`
}

// UserResponse レスポンス用のJSON形式を表した構造体
type UserResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"user_name"`
	Email string `json:"email"`
}

// UsersResponse Userリストレスポンス用のJSON形式を表した構造体
type UsersResponse struct {
	Users []*UserResponse `json:"users"`
}

// PostUsers 新規作成
func (ctrl *UserController) PostUsers(ctx *gin.Context) {
	ctrl.log.Info("Starting PostUsers handler")

	// リクエストボディを取得
	body, err := ctx.GetRawData()
	if err != nil {
		ctrl.log.Error("Failed to get request body", "error", err)
		Response500(ctx, err)
		return
	}

	// バリデーション処理
	validator := PostSettingValidator()
	validErr := validator.ValidateBody(string(body))
	if validErr != nil {
		ctrl.log.Warn("Validation failed", "errors", validErr)
		Response400(ctx, validErr)
		return
	}

	// JSON形式から構造体に変換
	var req RequestPostUser
	err = json.Unmarshal(body, &req)
	if err != nil {
		ctrl.log.Error("Failed to unmarshal request body", "error", err)
		Response500(ctx, err)
		return
	}

	// 新規作成処理
	ctrl.log.Info("Creating new user", "user_name", req.Name, "email", req.Email)
	creator := registry.GetFactory().BuildCreateUser()
	res, err := creator.Execute(&usecase.CreateUserRequest{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		if err.Error() == interactor.ErrUniqEmail.Error() {
			ctrl.log.Warn("Email already registered", "email", req.Email)
			Response400(ctx, map[string]error{
				"email": errors.New("すでに登録されているメールアドレスです。"),
			})
			return
		}
		ctrl.log.Error("Failed to create user", "error", err)
		Response500(ctx, err)
		return
	}

	ctrl.log.Info("User created successfully", "userID", res.GetUserID())
	// 201レスポンス
	Response201(ctx, res.GetUserID())
}

// PutUser 更新
func (ctrl *UserController) PutUser(ctx *gin.Context) {
	ctrl.log.Info("Starting PutUser handler")

	// リクエストボディを取得
	body, err := ctx.GetRawData()
	if err != nil {
		ctrl.log.Error("Failed to get request body", "error", err)
		Response500(ctx, err)
		return
	}

	// バリデーション処理
	validator := PostSettingValidator()
	validErr := validator.ValidateBody(string(body))
	if validErr != nil {
		ctrl.log.Warn("Validation failed", "errors", validErr)
		Response400(ctx, validErr)
		return
	}

	// JSON形式から構造体に変換
	var req RequestPutUser
	err = json.Unmarshal(body, &req)
	if err != nil {
		ctrl.log.Error("Failed to unmarshal request body", "error", err)
		Response500(ctx, err)
		return
	}

	// パスパラメータからユーザーIDを取得する
	userID, err := utils.ParseUint(ctx.Param("user_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse user_id", "error", err)
		Response500(ctx, err)
		return
	}

	// 更新処理
	ctrl.log.Info("Updating user", "userID", userID, "user_name", req.Name, "email", req.Email)
	updater := registry.GetFactory().BuildUpdateUser()
	_, err = updater.Execute(&usecase.UpdateUserRequest{
		ID:    userID,
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		if err.Error() == interactor.ErrUniqEmail.Error() {
			ctrl.log.Warn("Email already registered", "email", req.Email)
			Response400(ctx, map[string]error{
				"email": errors.New("すでに登録されているメールアドレスです。"),
			})
			return
		}
		ctrl.log.Error("Failed to update user", "error", err)
		Response500(ctx, err)
		return
	}

	ctrl.log.Info("User updated successfully", "userID", userID)

	// 200レスポンス
	Response200OK(ctx)
}

// GetUsers 一覧取得処理
func (ctrl *UserController) GetUsers(ctx *gin.Context) {
	ctrl.log.Info("Starting GetUsers handler")

	// 一覧取得処理
	getter := registry.GetFactory().BuildGetUserList()
	res, err := getter.Execute(&usecase.GetUserListRequest{})
	if err != nil {
		ctrl.log.Error("Failed to get user list", "error", err)
		Response500(ctx, err)
		return
	}

	// ドメインモデルからレスポンス用の構造体に詰め替える
	var resUsers = make([]*UserResponse, res.UserCount())
	for i, u := range res.Users {
		resUsers[i] = &UserResponse{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		}
	}

	ctrl.log.Info("User list retrieved successfully", "count", len(resUsers))
	// レスポンス処理
	Response200(ctx, &UsersResponse{
		Users: resUsers,
	})
}

// GetUser IDから取得
func (ctrl *UserController) GetUser(ctx *gin.Context) {
	ctrl.log.Info("Starting GetUser handler")

	// パスパラメータからユーザーIDを取得する
	userID, err := utils.ParseUint(ctx.Param("user_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse user_id", "error", err)
		Response500(ctx, err)
		return
	}

	// ユーザー取得処理
	ctrl.log.Info("Getting user by ID", "userID", userID)
	getter := registry.GetFactory().BuildGetUserByID()
	res, err := getter.Execute(&usecase.GetUserByIDRequest{UserID: userID})
	if err != nil {
		if err.Error() == domain.ErrNotFound.Error() {
			ctrl.log.Warn("User not found", "userID", userID)
			Response404(ctx)
			return
		}
		ctrl.log.Error("Failed to get user", "error", err)
		Response500(ctx, err)
		return
	}

	ctrl.log.Info("User retrieved successfully", "userID", res.User.ID)
	// ドメインモデルからレスポンス用構造体に詰め替えて、レスポンス
	Response200(ctx, &UserResponse{
		ID:    res.User.ID,
		Name:  res.User.Name,
		Email: res.User.Email,
	})
}

// DeleteUser 削除処理
func (ctrl *UserController) DeleteUser(ctx *gin.Context) {
	ctrl.log.Info("Starting DeleteUser handler")

	// パスパラメータからユーザーIDを取得する
	userID, err := utils.ParseUint(ctx.Param("user_id"))
	if err != nil {
		ctrl.log.Error("Failed to parse user_id", "error", err)
		Response500(ctx, err)
		return
	}

	// 削除処理
	ctrl.log.Info("Deleting user", "userID", userID)
	deleter := registry.GetFactory().BuildUserDeleter()
	_, err = deleter.Execute(&usecase.DeleteUserRequest{
		UserID: userID,
	})
	if err != nil {
		ctrl.log.Error("Failed to delete user", "error", err)
		Response500(ctx, err)
		return
	}

	ctrl.log.Info("User deleted successfully", "userID", userID)
	// レスポンス
	Response200OK(ctx)
}
