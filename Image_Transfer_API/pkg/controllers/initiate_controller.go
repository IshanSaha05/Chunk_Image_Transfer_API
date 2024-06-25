package controllers

import (
	"Image-Transfer-API/pkg/helpers"
	"Image-Transfer-API/pkg/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func InitiateUploadProcess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Creating a context with timeout.
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		// Validating the response data and checking timeout after validation.
		var responseData models.InitateData
		if !helpers.ValidateResponseData(c, &responseData) {
			c.Abort()
			return
		}
		if c.Err() != nil {
			if c.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
			}

			c.Abort()
			return
		}

		// Creates the session id, stores session and metadata info into redis and checking timeout after validation.
		helpers.SessionCreateSave(c, responseData)
		if c.Err() != nil {
			if c.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
			}

			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Successful creation of the session."})
	}
}
