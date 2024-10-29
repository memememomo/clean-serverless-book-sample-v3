package controller

import (
	"clean-serverless-book-sample/logger"

	"github.com/gin-gonic/gin"
)

func Routes() *gin.Engine {
	r := gin.Default()

	log := logger.GetLogger()

	userCtrl := &UserController{
		log: log,
	}
	r.POST("/v1/users", userCtrl.PostUsers)
	r.GET("/v1/users", userCtrl.GetUsers)
	r.GET("/v1/users/:user_id", userCtrl.GetUser)
	r.PUT("/v1/users/:user_id", userCtrl.PutUser)
	r.DELETE("/v1/users/:user_id", userCtrl.DeleteUser)

	micropostCtrl := &MicropostController{
		log: log,
	}
	r.POST("/v1/users/:user_id/microposts", micropostCtrl.PostMicroposts)
	r.GET("/v1/users/:user_id/microposts", micropostCtrl.GetMicroposts)
	r.GET("/v1/users/:user_id/microposts/:micropost_id", micropostCtrl.GetMicropost)
	r.PUT("/v1/users/:user_id/microposts/:micropost_id", micropostCtrl.PutMicropost)
	r.DELETE("/v1/users/:user_id/microposts/:micropost_id", micropostCtrl.DeleteMicropost)
	return r
}
