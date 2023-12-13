package hdfs_store

import (
	"fmt"
	"github.com/colinmarc/hdfs/v2"
	"log"
	"os"
	"strings"
)

// NoSQL: FileStorage struct encapsulating HDFS client
type FileStorage struct {
	client *hdfs.Client
	logger *log.Logger
}

func New(logger *log.Logger) (*FileStorage, error) {
	// Local instance
	hdfsUri := os.Getenv("HDFS_URI")

	client, err := hdfs.New(hdfsUri)
	if err != nil {
		logger.Panic(err)
		return nil, err
	}

	// Return storage handler with logger and HDFS client
	return &FileStorage{
		client: client,
		logger: logger,
	}, nil
}

func (fs *FileStorage) Close() {
	// Close all underlying connections to the HDFS server
	fs.client.Close()
}

func (fs *FileStorage) CreateDirectories() error {
	// Default permissions
	// 0644 Only the owner can read and write. Everyone else can only read. No one can execute the file.
	err := fs.client.MkdirAll(hdfsCopyDir, 0644)
	if err != nil {
		fs.logger.Println(err)
		return err
	}

	err = fs.client.Mkdir(hdfsWriteDir, 0644)
	if err != nil {
		fs.logger.Println(err)
		return err
	}

	return nil
}

func (fs *FileStorage) WalkDirectories() []string {
	// Walk all files in HDFS root directory and all subdirectories
	var paths []string
	callbackFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			fs.logger.Printf("Directory: %s\n", path)
			path = fmt.Sprintf("Directory: %s\n", path)
			paths = append(paths, path)
		} else {
			fs.logger.Printf("File: %s\n", path)
			path = fmt.Sprintf("File: %s\n", path)
			paths = append(paths, path)
		}
		return nil
	}
	fs.client.Walk(hdfsRoot, callbackFunc)
	return paths
}

func (fs *FileStorage) CopyLocalFile(localFilePath, fileName string) error {
	// Create local file
	file, err := os.Create(localFilePath)
	if err != nil {
		fs.logger.Println("Error in creating local file:", err)
		return err
	}
	fileContent := "Hello World!"
	_, err = file.WriteString(fileContent)
	if err != nil {
		fs.logger.Println("Error in writing local file:", err)
		return err
	}
	file.Close()

	// Copy file to HDFS
	_ = fs.client.CopyToRemote(localFilePath, hdfsCopyDir+fileName)
	return nil
}

func (fs *FileStorage) WriteFile(fileContent string, fileName string) error {
	filePath := hdfsWriteDir + fileName

	// Create file on HDFS with default replication and block size
	file, err := fs.client.Create(filePath)
	if err != nil {
		fs.logger.Println("Error in creating file on HDFS:", err)
		return err
	}

	// Write content
	// Create byte array from string file content
	fileContentByteArray := []byte(fileContent)

	// IMPORTANT: writes content to local buffer, content is pushed to HDFS only when Close is called!
	_, err = file.Write(fileContentByteArray)
	if err != nil {
		fs.logger.Println("Error in writing file on HDFS:", err)
		return err
	}

	// CLOSE FILE WHEN ALL WRITING IS DONE
	// Ensuring all changes are flushed to HDFS
	// defer file.Close() can be used at the begining of the method to ensure closing is not forgotten
	_ = file.Close()
	return nil
}

func (fs *FileStorage) ReadFile(fileName string, isCopied bool) (string, error) {
	var filePath string
	if isCopied {
		filePath = hdfsCopyDir + fileName
	} else {
		filePath = hdfsWriteDir + fileName
	}

	// Open file for reading
	file, err := fs.client.Open(filePath)
	if err != nil {
		fs.logger.Println("Error in opening file for reding on HDFS:", err)
		return "", err
	}

	// Read file content
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil {
		fs.logger.Println("Error in reading file on HDFS:", err)
		return "", err
	}

	// Convert number of read bytes into string
	fileContent := string(buffer[:n])
	return fileContent, nil
}

func extractFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func (fs *FileStorage) StoreImageInHDFS(imageURL string) (string, error) {
	fileName := extractFileNameFromURL(imageURL)

	// Create local file
	localFilePath := "/tmp/" + fileName // Adjust the local path as needed
	err := fs.CopyLocalFile(localFilePath, fileName)
	if err != nil {
		return "", fmt.Errorf("error copying local file: %v", err)
	}

	// Copy file to HDFS
	hdfsFilePath := hdfsCopyDir + fileName
	err = fs.client.CopyToRemote(localFilePath, hdfsFilePath)
	if err != nil {
		return "", fmt.Errorf("error copying file to HDFS: %v", err)
	}

	return hdfsFilePath, nil
}
