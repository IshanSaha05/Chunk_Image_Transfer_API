package routes

import (
	"Image-Transfer-API/pkg/controllers"

	"github.com/gin-gonic/gin"
)

func InitiateRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/upload/initiate", controllers.InitiateUploadProcess())
}
