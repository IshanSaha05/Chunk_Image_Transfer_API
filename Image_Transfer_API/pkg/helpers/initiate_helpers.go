package helpers

import (
	"Image-Transfer-API/pkg/models"
	redis_database "Image-Transfer-API/pkg/redis"

	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ValidateResponseData(c *gin.Context, responseData *models.InitateData) bool {
	// Initiating the reposne data to bind it into from json format to declared struct.
	if err := c.BindJSON(responseData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}

	if c.Err() != nil {
		if c.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
		}
		return false
	}

	// Validating the response data.
	models.CustomValidators()
	err := models.Validate.Struct(responseData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}

	return true
}

// Function to create new session store it in a cookie and also save the session in redis.
func SessionCreateSave(c *gin.Context, responseData models.InitateData) *http.Cookie {
	// Creating new session id.
	sessionID := uuid.NewString()
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout."})
		return nil
	}

	// Creating the session struct to store it in redis.
	session := models.Session{
		SessionID: sessionID,
		UserID:    "Guest",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		FileData:  responseData,
	}
	if c.Err() != nil {
		if c.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
		}
		return nil
	}

	// Getting the redis client.
	redisClient := redis_database.GetRedisClient()
	if c.Err() != nil {
		if c.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
		}
		return nil
	}

	// Set the key-value pair with sessionId as the key and session struct as the value and expiration time of the key-value pair as 1 hr.
	err := redisClient.SessionClient.Set(*redisClient.Ctx, sessionID, session, 24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error})
		return nil
	}
	if c.Err() != nil {
		if c.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
		}
		return nil
	}

	// Set the key-value pair to keep in track which chunks has been received.
	for i := 1; i <= responseData.TotalChunks; i++ {
		chunkInfo := models.ChunkInfo{
			ChunkNumber: i,
			IsReceived:  false,
		}
		err := redisClient.ChunksClient.Set(*redisClient.Ctx, fmt.Sprintf("%s__%d", sessionID, i), chunkInfo, 0).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error})
			return nil
		}
	}
	if c.Err() != nil {
		if c.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
		}
		return nil
	}

	// Create a cookie.
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
	}

	// Setting the cookie in the header.
	http.SetCookie(c.Writer, cookie)

	return cookie
}
