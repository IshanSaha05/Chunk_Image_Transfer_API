package models

import (
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Structure for taking in data send in response during upload process initiation.
type InitateData struct {
	FileType     string `json:"fileType" validate:"required,filetype"`
	FileSize     int    `json:"fileSize" validate:"required,maxfilesize"`
	FileSizeUnit string `json:"fileSizeUnit" validate:"required,filesizeunit"`
	FileName     string `json:"fileName" validate:"required,filename"`
	TotalChunks  int    `json:"totalChunks" validate:"required,lt=1000"`
}

var Validate *validator.Validate

// Function to validate the file types.
func fileTypeValidator(fl validator.FieldLevel) bool {
	fileType := fl.Field().String()
	allowedTypes := []string{"jpg", "jpeg", "png"}
	for _, t := range allowedTypes {
		if strings.ToLower(fileType) == t {
			return true
		}
	}
	return false
}

// Function to validate the file size type.
func fileSizeUnitValidator(fl validator.FieldLevel) bool {
	fileSizeUnit := fl.Field().String()
	allowedUnits := []string{"mb", "gb", "kb"}
	for _, unit := range allowedUnits {
		if strings.ToLower(fileSizeUnit) == unit {
			return true
		}
	}
	return false
}

// Function to validate the maximum file size.
func maxFileSizeValidator(fl validator.FieldLevel) bool {
	fileSize := fl.Parent().FieldByName("FileSize").Int()
	fileSizeUnit := strings.ToLower(fl.Parent().FieldByName("FileSizeUnit").String())

	switch fileSizeUnit {
	case "mb":
		return fileSize <= 100
	case "gb":
		return fileSize <= 0
	case "kb":
		return fileSize <= 100000
	default:
		return false
	}
}

// Function to validate the file name, if it follows the proper format.
func fileNameValidator(fl validator.FieldLevel) bool {
	fileName := fl.Field().String()
	regex := `^[a-zA-Z0-9._-]+$`
	match, _ := regexp.MatchString(regex, fileName)
	return match && !strings.HasPrefix(fileName, "-")
}

// Function to register the custom validators
func CustomValidators() {
	Validate.RegisterValidation("filetype", fileTypeValidator)
	Validate.RegisterValidation("filesizeunit", fileSizeUnitValidator)
	Validate.RegisterValidation("maxfilesize", maxFileSizeValidator)
	Validate.RegisterValidation("filename", fileNameValidator)
}

// Structure for storing session details.
type Session struct {
	SessionID string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
	FileData  InitateData
}

// To store chunk info about a particular file of a particular session.
type ChunkInfo struct {
	ChunkNumber int
	IsReceived  bool
}
