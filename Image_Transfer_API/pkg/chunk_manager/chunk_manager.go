package chunkmanager

import (
	"Image-Transfer-API/pkg/models"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func MergeChunks(session models.Session) error {
	// Getting the temporary folder path.
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("error while loading environment variables")
	}

	tempFolderPath := os.Getenv("TEMP_UPLOAD_PATH")
	if tempFolderPath == "" {
		return fmt.Errorf("folder path not set")
	}

	tempFolderPath = filepath.Join(tempFolderPath, "."+session.SessionID)

	// Getting the final folder path and file path.
	finalFolderPath := os.Getenv("FINAL_UPLOAD_PATH")
	if finalFolderPath == "" {
		return fmt.Errorf("folder path not set")
	}
	finalFolderPath = filepath.Join(finalFolderPath, session.SessionID)
	finalFilePath := filepath.Join(finalFolderPath, fmt.Sprintf("%s.%s", session.SessionID, session.FileData.FileType))

	// Opening the final output file for writing.
	finalFile, err := os.Create(finalFilePath)
	if err != nil {
		return fmt.Errorf("problem while creating the final file")
	}
	defer finalFile.Close()

	// Reading each chunk file and writing it to the output file.
	for i := 1; i <= session.FileData.TotalChunks; i++ {
		// Getting the chunk file name and chunk file path.
		chunkFileName := fmt.Sprintf("%s_%d.%s", session.SessionID, i, session.FileData.FileType)
		chunkFilePath := filepath.Join(tempFolderPath, chunkFileName)

		// Reading the chunk file.
		chunkData, err := os.ReadFile(chunkFilePath)
		if err != nil {
			return fmt.Errorf("problem while reading the chunk file as part of merging into the final file")
		}

		// Writing the chunk file into the final file finally.
		_, err = finalFile.Write(chunkData)
		if err != nil {
			return fmt.Errorf("problem while writing the chunk file as part of merging into the final file")
		}
	}

	// Removing the temporary folder and its contents.
	err = os.RemoveAll(tempFolderPath)
	if err != nil {
		return fmt.Errorf("problem while deleting the temporary folder for the session: %s", session.SessionID)
	}

	return nil
}
