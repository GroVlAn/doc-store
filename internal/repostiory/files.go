package repository

import (
	"fmt"
	"os"
)

const (
	filesDirectory = "/files"
)

type FileRepository struct{}

func NewFileRepository() *FileRepository {
	createFilesDirectory(filesDirectory)

	return &FileRepository{}
}

func (fr *FileRepository) SaveFile(userID string, fileName string, file []byte) error {
	createFilesDirectory(fmt.Sprintf("%s/%s", filesDirectory, userID))

	filePath := fmt.Sprintf("%s/%s/%s", filesDirectory, userID, fileName)
	err := os.WriteFile(filePath, file, os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}

func (fr *FileRepository) File(userID string, fileName string) (string, error) {
	filePath := fmt.Sprintf("%s/%s/%s", filesDirectory, userID, fileName)

	return filePath, nil
}

func (fr *FileRepository) DeleteFile(userID string, fileName string) error {
	filePath := fmt.Sprintf("%s/%s/%s", filesDirectory, userID, fileName)

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("removing file: %w", err)
	}

	return nil
}

func (fr *FileRepository) FileExist(userID string, fileName string) bool {
	filePath := fmt.Sprintf("%s/%s/%s", filesDirectory, userID, fileName)

	return pathExist(filePath)
}

func createFilesDirectory(dirPath string) {
	if pathExist(dirPath) {
		return
	}

	_ = os.Mkdir(dirPath, os.ModePerm)
}

func pathExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}
