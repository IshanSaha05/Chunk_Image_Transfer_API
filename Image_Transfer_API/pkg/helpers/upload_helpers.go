package helpers

import (
	"Image-Transfer-API/pkg/models"
	redis_database "Image-Transfer-API/pkg/redis"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func CheckFileLocation(c *gin.Context, session models.Session) (bool, string) {
	// Get the upload location path from the environment file.
	err := godotenv.Load(".env")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Error while loading environment variables."})
		return false, ""
	}

	folderPath := os.Getenv("TEMP_UPLOAD_PATH")
	if folderPath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Folder path not set."})
		return false, ""
	}
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		return false, ""
	}

	// Update the temp location with the session name and make a folder if does not exists.
	folderPath = filepath.Join(folderPath, "."+session.SessionID)

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while making chunk folder path."})
			return false, ""
		}
	}
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		return false, ""
	}

	return true, folderPath
}

func ValidateChunks(c *gin.Context, session models.Session) (bool, *multipart.FileHeader) {
	// Get the environment variables for max upload size.
	err := godotenv.Load(".env")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Error while loading environment variables."})
		return false, nil
	}

	maxUploadSizeString := os.Getenv("MAX_UPLOAD_SIZE")
	if maxUploadSizeString == "" {
		maxUploadSizeString = "10485760"
	}
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		return false, nil
	}

	// Convert the max size from string to integer.
	maxUploadSize, err := strconv.Atoi(maxUploadSizeString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting max size from string to int"})
		return false, nil
	}
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		return false, nil
	}

	// Wrap the request body with a condition to accept only file size less than or equal to 10 >> 20 bytes size.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxUploadSize))
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error: No file received."})
		return false, nil
	}
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		return false, nil
	}

	// Getting the chunk extension from the chunk file shared in the request.
	chunkExt := strings.ToLower(filepath.Ext(file.Filename))

	// Getting the file extension of the total file shared during session creation.
	fileExt := session.FileData.FileType

	// Check whether the file shared is of accepted image type or not, by comparing it with the shared file chunk type
	//and already stored file type in the storage which was shared during session creation.
	if chunkExt != fileExt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error: File extension of the chunk file is not same."})
	}
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		return false, nil
	}

	return true, file
}

func SaveChunk(c *gin.Context, chunkNumber int, folderPath string, file *multipart.FileHeader, session models.Session) bool {
	// Making the file name with the session id and chunknumber and finally making the absolute file path.
	fileName := fmt.Sprintf("%s_%d.%s", session.SessionID, chunkNumber, session.FileData.FileType)
	filePath := filepath.Join(folderPath, fileName)

	// Saving the file in the temporary location.
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save file."})
		return false
	}
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		return false
	}

	return true
}

func UpdateStatusRedis(c *gin.Context, session models.Session, chunkNumber int) bool {
	// Getting the redis client.
	redisClient := redis_database.GetRedisClient()
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		return false
	}

	// Checking whether the chunk line item for that session is present or not.
	_, err := redisClient.ChunksClient.Get(c, fmt.Sprintf("%s__%d", session.SessionID, chunkNumber)).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Chunk status line item not present in the redis."})
		return false
	}
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		return false
	}

	// If yes change the corresponding value from false to true.
	chunkInfo := models.ChunkInfo{
		ChunkNumber: chunkNumber,
		IsReceived:  true,
	}
	err = redisClient.ChunksClient.Set(*redisClient.Ctx, fmt.Sprintf("%s__%d", session.SessionID, chunkNumber), chunkInfo, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error})
		return false
	}
	if c.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		return false
	}

	return true
}
