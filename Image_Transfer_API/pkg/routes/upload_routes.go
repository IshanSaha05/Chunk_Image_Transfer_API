package routes

import (
	"Image-Transfer-API/pkg/controllers"
	"Image-Transfer-API/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func UploadRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.POST("/upload/chunk", controllers.UploadChunks())
}
