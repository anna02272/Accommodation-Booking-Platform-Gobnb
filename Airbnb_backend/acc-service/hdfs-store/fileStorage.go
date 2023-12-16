package hdfs_store

import (
	"acc-service/cache"
	"fmt"
	"github.com/colinmarc/hdfs/v2"
	"io"
	"log"
	"os"
	"strings"
)

// NoSQL: FileStorage struct encapsulating HDFS client
type FileStorage struct {
	Client *hdfs.Client
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
		Client: client,
		logger: logger,
	}, nil
}

func (fs *FileStorage) Close() {
	// Close all underlying connections to the HDFS server
	fs.Client.Close()
}

func (fs *FileStorage) CreateDirectories() error {
	// Default permissions
	err := fs.Client.MkdirAll(hdfsCopyDir, 0644)
	if err != nil {
		fs.logger.Println(err)
		return err
	}

	err = fs.Client.Mkdir(hdfsWriteDir, 0644)
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
	fs.Client.Walk(hdfsRoot, callbackFunc)
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
	_ = fs.Client.CopyToRemote(localFilePath, hdfsCopyDir+fileName)
	return nil
}

func (fs *FileStorage) WriteFile(fileContent string, fileName string) error {
	filePath := hdfsWriteDir + fileName

	// Create file on HDFS with default replication and block size
	file, err := fs.Client.Create(filePath)
	if err != nil {
		fs.logger.Println("Error in creating file on HDFS:", err)
		return err
	}

	// Write content
	// Create byte array from string file content
	fileContentByteArray := []byte(fileContent)

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
	file, err := fs.Client.Open(filePath)
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

func (fs *FileStorage) StoreImageInHDFS(imageData []byte) (string, error) {
	fileName := cache.GenerateUniqueImageID()

	localFilePath := "/tmp/" + fileName
	err := fs.WriteLocalFile(localFilePath, imageData)
	if err != nil {
		fs.logger.Printf("Error writing local file: %v\n", err)
		return "", fmt.Errorf("error writing local file: %v", err)
	}

	hdfsFilePath := hdfsCopyDir + fileName
	err = fs.Client.CopyToRemote(localFilePath, hdfsFilePath)
	if err != nil {
		fs.logger.Printf("Error copying file to HDFS: %v\n", err)
		return "", fmt.Errorf("error copying file to HDFS: %v", err)
	}

	fs.logger.Printf("Image successfully stored in HDFS. HDFS Path: %s\n", hdfsFilePath)
	fmt.Println("HDFS file path")
	fmt.Println(hdfsFilePath)

	return hdfsFilePath, nil
}

func (fs *FileStorage) WriteLocalFile(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		fs.logger.Printf("Error creating local file: %v\n", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		fs.logger.Printf("Error writing local file: %v\n", err)
		return err
	}

	return nil
}

func (fs *FileStorage) WriteFileBytes(imageData []byte, fileName string, dirName string) error {
	filePath := hdfsWriteDir + fileName
	fmt.Println(filePath)
	fmt.Println("here filePath")

	// If the directory doesn't exist, create it
	err := fs.Client.Mkdir(hdfsWriteDir+dirName, 0777) // You can adjust the permission mode as needed
	if err != nil {
		fmt.Println(err)
		fmt.Println("Ok")
	}

	file, err := fs.Client.Create(filePath + ".jpg")
	if err != nil {
		fs.logger.Println("Error in creating file on HDFS:", err)
		return err
	}

	_, err = file.Write(imageData)
	if err != nil {
		fs.logger.Println("Error in writing image to file on HDFS:", err)
		return err
	}

	err = file.Close()
	if err != nil {
		fs.logger.Println("Error in closing file on HDFS:", err)
		return err
	}

	return nil
}

func (fs *FileStorage) ReadFileBytes(fileName string) ([]byte, error) {
	file, err := fs.Client.Open(fileName)
	if err != nil {
		fs.logger.Println("Error in opening file for reading on HDFS:", err)
		return nil, err
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		fs.logger.Println("Error in reading file on HDFS:", err)
		return nil, err
	}

	return fileContent, nil
}
