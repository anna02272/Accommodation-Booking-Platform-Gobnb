package hdfs_store

import (
	"acc-service/cache"
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

func (fs *FileStorage) StoreImageInHDFS(imageData []byte) (string, error) {
	fileName := cache.GenerateUniqueImageID()

	localFilePath := "/tmp/" + fileName
	err := fs.WriteLocalFile(localFilePath, imageData)
	if err != nil {
		fs.logger.Printf("Error writing local file: %v\n", err)
		return "", fmt.Errorf("error writing local file: %v", err)
	}

	hdfsFilePath := hdfsCopyDir + fileName
	err = fs.client.CopyToRemote(localFilePath, hdfsFilePath)
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

//func (fs *FileStorage) GetFirstImageURL(accommodationID string) (string, error) {
//	// Construct the cache key for the first image of the accommodation
//	cacheKey := fmt.Sprintf(cacheAll, accommodationID)
//
//	// Check if the image URL is already cached
//	imageURL, err := fs.cli.Get(cacheKey).Result()
//	if err == nil {
//		return imageURL, nil
//	}
//
//	// If not cached, fetch the image data from HDFS or cache and construct the URL
//	// Add your logic to fetch the image data from HDFS or cache
//	// For example, you might have a method like GetImageFromHDFS(accommodationID) in your code
//	imageData, err := fs.GetImageFromHDFS(accommodationID)
//	if err != nil {
//		return "", err
//	}
//
//	// Cache the image URL for future use
//	fs.cli.Set(cacheKey, imageData, 300*time.Second)
//
//	// Construct the URL based on the accommodation ID and file path
//	fileName := cache.GenerateUniqueImageID()
//	hdfsFilePath := hdfsCopyDir + fileName
//	imageURL = "https://your-hdfs-server" + hdfsFilePath
//
//	return imageURL, nil
//}
//
//// Example method to fetch image data from HDFS based on accommodation ID
//func (fs *FileStorage) GetImageFromHDFS(accommodationID string) (string, error) {
//	// Add your logic to fetch the image data from HDFS
//	// For example, you might have a method like ReadFileFromHDFS(accommodationID) in your code
//	// Ensure to handle errors and return the image data as a string
//	imageData, err := fs.ReadFileFromHDFS(accommodationID)
//	if err != nil {
//		return "", err
//	}
//
//	return imageData, nil
//}
