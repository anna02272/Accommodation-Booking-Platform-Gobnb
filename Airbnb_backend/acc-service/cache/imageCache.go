package cache

import (
	"encoding/base64"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"os"
	"time"
)

const (
	cacheImages = "images:%s:%s"
	cacheAll    = "images"
)

type ImageCache struct {
	cli    *redis.Client
	logger *log.Logger
}

// Construct Redis client
func New(logger *log.Logger) *ImageCache {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisAddress := fmt.Sprintf("%s:%s", redisHost, redisPort)

	client := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	return &ImageCache{
		cli:    client,
		logger: logger,
	}
}

func (pc *ImageCache) Ping() {
	val, _ := pc.cli.Ping().Result()
	pc.logger.Println(val)
}

func (ic *ImageCache) PostImage(imageID string, accID string, imageData []byte) error {
	key := constructImageKey(imageID, accID)

	encodedImage := base64.StdEncoding.EncodeToString(imageData)

	err := ic.cli.Set(key, encodedImage, 300*time.Second).Err()
	if err != nil {
		fmt.Println("Error setting image in Redis:", err)
		return err
	}
	return err
}

func (ic *ImageCache) GetImage(imageID, accID string) ([]byte, error) {
	key := constructImageKey(imageID, accID)
	imageData, err := ic.cli.Get(key).Bytes()
	if err != nil {
		return nil, err
	}
	ic.logger.Println("Image Cache hit")
	return imageData, nil
}

func (ic *ImageCache) ImageExists(imageID, accID string) bool {
	key := constructImageKey(imageID, accID)
	cnt, err := ic.cli.Exists(key).Result()
	if cnt == 1 {
		return true
	}
	if err != nil {
		return false
	}
	return false
}

func (ic *ImageCache) CacheImage(imageID, accID string, imageData string) error {
	expiration := 30 * time.Second
	key := constructImageKey(imageID, accID)
	err := ic.cli.Set(key, imageData, expiration).Err()
	return err
}

// Helper function to construct image cache key
func constructImageKey(imageID, accID string) string {
	return fmt.Sprintf(cacheImages, accID, imageID)
}

func GenerateUniqueImageID() string {
	return fmt.Sprintf("image_%d", time.Now().UnixNano())
}
