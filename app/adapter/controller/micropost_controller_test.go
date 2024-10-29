package controller

import (
	"bytes"
	"clean-serverless-book-sample/domain"
	"clean-serverless-book-sample/mocks"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/k0kubun/pp"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return Routes()
}

// TestPostMicroposts_201 新規作成処理 正常時
func TestPostMicroposts_201(t *testing.T) {
	// テスト用DynamoDBの設定
	tables := mocks.SetupDB(t)
	defer tables.Cleanup()

	router := setupRouter()

	// リクエスト用パラメータ
	body := map[string]interface{}{
		"content": strings.Repeat("a", 140),
	}
	bodyBytes, err := json.Marshal(body)
	assert.NoError(t, err)

	userID := uint64(1)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/users/%d/microposts", userID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// レスポンスコードチェック
	assert.Equal(t, 201, w.Code)

	// JSONからmap型に変換
	var resBody map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resBody)
	assert.NoError(t, err)

	// 新規作成されたIDの値をチェック
	assert.Equal(t, "1", resBody["id"])

	// DynamoDBに保存されているかチェック
	micropost, err := tables.MicropostOperator.GetMicropostByID(1)
	assert.NoError(t, err)
	assert.Equal(t, body["content"].(string), micropost.Content)
	assert.Equal(t, userID, micropost.UserID)
}

// TestPostMicroposts_400 新規作成処理 バリデーションエラー時
func TestPostMicroposts_400(t *testing.T) {
	// テスト用DynamoDB設定
	tables := mocks.SetupDB(t)
	defer tables.Cleanup()

	router := setupRouter()

	cases := []struct {
		Request  map[string]interface{}
		Expected map[string]interface{}
	}{
		// 未入力の場合
		{
			Request: map[string]interface{}{
				"content": "",
			},
			Expected: map[string]interface{}{
				"content": "本文を入力してください。",
			},
		},
		// 本文の文字数が上限を超えている場合
		{
			Request: map[string]interface{}{
				"content": strings.Repeat("a", 141),
			},
			Expected: map[string]interface{}{
				"content": "本文の文字数が上限を超えています。",
			},
		},
	}

	for i, c := range cases {
		msg := fmt.Sprintf("Case:%d", i+1)

		bodyBytes, err := json.Marshal(c.Request)
		assert.NoError(t, err)

		req, _ := http.NewRequest("POST", "/v1/users/1/microposts", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		pp.Println(string(w.Body.Bytes()))
		var resBody map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.NoError(t, err)

		errors := resBody["errors"].(map[string]interface{})

		assert.Equal(t, 400, w.Code, msg)
		assert.Equal(t, c.Expected, errors)
	}
}

// TestPutMicropost_200 更新処理 正常時
func TestPutMicropost_200(t *testing.T) {
	// テスト用のDynamoDBを設定
	tables := mocks.SetupDB(t)
	defer tables.Cleanup()

	router := setupRouter()

	// 更新用モックデータを作成
	micropostMock, err := tables.MicropostOperator.CreateMicropost(&domain.MicropostModel{
		Content: "Content_1",
		UserID:  1,
	})
	assert.NoError(t, err)

	// 更新用リクエスト
	body := map[string]interface{}{
		"content": strings.Repeat("a", 140),
	}
	bodyBytes, err := json.Marshal(body)
	assert.NoError(t, err)

	req, _ := http.NewRequest("PUT",
		fmt.Sprintf("/v1/users/%d/microposts/%d", micropostMock.UserID, micropostMock.ID),
		bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// レスポンスコードをチェック
	assert.Equal(t, 200, w.Code)

	// DynamoDBに更新データが反映されているかチェック
	micropost, err := tables.MicropostOperator.GetMicropostByID(micropostMock.ID)
	assert.NoError(t, err)
	assert.Equal(t, body["content"].(string), micropost.Content)
}

// TestGetMicropost 取得処理
func TestGetMicropost(t *testing.T) {
	// テスト用のDynamoDBを設定
	tables := mocks.SetupDB(t)
	defer tables.Cleanup()

	router := setupRouter()

	// 取得用のモックデータを作成
	micropostMock, err := tables.MicropostOperator.CreateMicropost(&domain.MicropostModel{
		Content: "Content_1",
		UserID:  1,
	})
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET",
		fmt.Sprintf("/v1/users/%d/microposts/%d", micropostMock.UserID, micropostMock.ID),
		nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// レスポンスコードをチェック
	assert.Equal(t, 200, w.Code)

	// 取得したデータをチェック
	var body map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, float64(micropostMock.ID), body["id"])
	assert.Equal(t, micropostMock.Content, body["content"])
	assert.Equal(t, float64(micropostMock.UserID), body["user_id"])
}

// TestGetMicroposts 一覧取得処理
func TestGetMicroposts(t *testing.T) {
	tables := mocks.SetupDB(t)
	defer tables.Cleanup()

	router := setupRouter()

	// 取得用のモックデータを作成
	micropostMock1, err := tables.MicropostOperator.CreateMicropost(&domain.MicropostModel{
		Content: "Content_1",
		UserID:  1,
	})
	assert.NoError(t, err)

	micropostMock2, err := tables.MicropostOperator.CreateMicropost(&domain.MicropostModel{
		Content: "Content_2",
		UserID:  1,
	})
	assert.NoError(t, err)

	// このデータはUserIDが異なるので取得されない想定
	_, err = tables.MicropostOperator.CreateMicropost(&domain.MicropostModel{
		Content: "Content_3",
		UserID:  2,
	})
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/v1/users/1/microposts", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// レスポンスコードをチェック
	assert.Equal(t, 200, w.Code)

	// JSONからmap型に変換
	var body map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)

	// 取得したデータをチェック
	actualMicroposts := body["microposts"].([]interface{})

	expected1 := micropostMock1
	actual1 := actualMicroposts[0].(map[string]interface{})
	assert.Equal(t, float64(expected1.ID), actual1["id"])
	assert.Equal(t, expected1.Content, actual1["content"])
	assert.Equal(t, float64(expected1.UserID), actual1["user_id"])

	expected2 := micropostMock2
	actual2 := actualMicroposts[1].(map[string]interface{})
	assert.Equal(t, float64(expected2.ID), actual2["id"])
	assert.Equal(t, expected2.Content, actual2["content"])
	assert.Equal(t, float64(expected2.UserID), actual2["user_id"])
}

// TestDeleteMicropost 削除処理
func TestDeleteMicropost(t *testing.T) {
	// テスト用のDynamoDBを設定
	tables := mocks.SetupDB(t)
	defer tables.Cleanup()

	router := setupRouter()

	// 削除用モックデータを作成
	micropostMock, err := tables.MicropostOperator.CreateMicropost(&domain.MicropostModel{
		Content: "Content_1",
		UserID:  1,
	})
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE",
		fmt.Sprintf("/v1/users/%d/microposts/%d", micropostMock.UserID, micropostMock.ID),
		nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// ステータスコードをチェック
	assert.Equal(t, 200, w.Code)

	// DynamoDBからデータが削除されているかチェック
	microposts, err := tables.MicropostOperator.GetMicropostsByUserID(micropostMock.UserID)
	assert.NoError(t, err)
	assert.Len(t, microposts, 0)
}
