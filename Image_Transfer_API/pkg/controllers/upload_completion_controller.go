package controllers

import (
	chunkmanager "Image-Transfer-API/pkg/chunk_manager"
	"Image-Transfer-API/pkg/helpers"
	"Image-Transfer-API/pkg/models"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadCompletion() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Creating a context with timeout of 100 seconds and adding it the gin contxt.
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// Getting the session data from the middleware.
		temp, exists := c.Get("session_data")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: No data passed from middleware to main handler."})
		}

		var session models.Session
		session, ok := temp.(models.Session)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting session from any type to defined struct."})
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

		// Checks the database if all the files have been uploaded or not.
		var responseData models.ChunkResponseInfo
		err := c.ShouldBindBodyWithJSON(&responseData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while de-serializing response data."})
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

		status := helpers.CheckReceivedStatus(c, responseData.ChunkNumber, session)

		// Run the chunk manager and create the final file.
		err = chunkmanager.MergeChunks(session)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: %s", err.Error())})
			c.Abort()
			return
		}

		// If upload of all chunks and merger is successful, return a proper json.
		if status {
			c.JSON(http.StatusOK, gin.H{"message": "Message: All chunks are successfully uploaded."})
		}
	}
}
