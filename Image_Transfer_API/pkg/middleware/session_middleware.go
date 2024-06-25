package middleware

import (
	"Image-Transfer-API/pkg/models"
	rediss "Image-Transfer-API/pkg/redis"
	"bytes"
	"context"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Creating a context with timeout for 100 seconds and adding it to the request context.
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// Gets the session id from the cookie sent in the client request.
		sessionId, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Error: No session cookie found."})
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

		// Get the redis client.
		redisClient := rediss.GetRedisClient()
		if c.Err() != nil {
			if c.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
			}

			c.Abort()
			return
		}

		// Get the stored session details from the redis if exists.
		val, err := redisClient.SessionClient.Get(*redisClient.Ctx, sessionId).Bytes()
		if c.Err() != nil {
			if c.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
			}

			c.Abort()
			return
		}

		// If session has been expired or it is invalid, the key with this session will not be present in the redis and will return redis.Nil.
		if err == redis.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Error: Session expired or invalid."})
			c.Abort()
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Failed to check session."})
			c.Abort()
			return
		}

		// De-serializing the data into the session struct.
		var session models.Session
		dec := gob.NewDecoder(bytes.NewReader(val))
		err = dec.Decode(&session)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Failed to decode session."})

			c.Abort()
			return
		}

		// Setting the session data for future use.
		c.Set("session_data", session)
		c.Next()
	}
}
