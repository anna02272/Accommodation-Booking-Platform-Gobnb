package cache

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"os"
	"time"
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

func (ic *ImageCache) PostImage(imageID string, imageData []byte) error {
	key := constructImageKey(imageID)
	err := ic.cli.Set(key, imageData, 30*time.Second).Err()
	return err
}

func (ic *ImageCache) GetImage(imageID string) ([]byte, error) {
	key := constructImageKey(imageID)
	imageData, err := ic.cli.Get(key).Bytes()
	if err != nil {
		return nil, err
	}
	ic.logger.Println("Image Cache hit")
	return imageData, nil
}

func (ic *ImageCache) ImageExists(imageID string) bool {
	key := constructImageKey(imageID)
	cnt, err := ic.cli.Exists(key).Result()
	if cnt == 1 {
		return true
	}
	if err != nil {
		return false
	}
	return false
}

func (ic *ImageCache) CacheImage(imageData string) error {
	expiration := 30 * time.Second
	imageID := generateUniqueImageID()
	key := constructImageKey(imageID)
	err := ic.cli.Set(key, imageData, expiration).Err()
	return err
}

// Helper function to construct image cache key
func constructImageKey(imageID string) string {
	return fmt.Sprintf("image:%s", imageID)
}

func generateUniqueImageID() string {
	return fmt.Sprintf("image_%d", time.Now().UnixNano())
}
