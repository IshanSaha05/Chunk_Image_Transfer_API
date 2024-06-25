package controllers

import (
	"Image-Transfer-API/pkg/helpers"
	"Image-Transfer-API/pkg/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadChunks() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Making a context with timeout and adding it to the gin context.
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

		// Checks whether the chunk upload location exists or not, if not makes one.
		status, location := helpers.CheckFileLocation(c, session)
		if !status {
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

		// Validates the chunk.
		status, file := helpers.ValidateChunks(c, session)
		if !status {
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

		// Getting the chunk number.
		var responseData models.ChunkResponseInfo
		if err := c.ShouldBindBodyWithJSON(&responseData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while binding json."})
			c.Abort()
			return
		}

		// Temporarily stores the chunk with a unique identifier.
		status = helpers.SaveChunk(c, responseData.ChunkNumber, location, file, session)
		if !status {
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

		// Updates the chunk status in the redis.
		status = helpers.UpdateStatusRedis(c, session, responseData.ChunkNumber)
		if !status {
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

		// Returns successful message after completion in json.
		c.JSON(http.StatusCreated, gin.H{"message": "File Uploaded Successfully",
			"filepath": location})

	}
}
