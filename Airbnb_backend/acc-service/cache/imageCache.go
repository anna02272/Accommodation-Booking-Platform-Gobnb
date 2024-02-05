package cache

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/go-redis/redis"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	Tracer trace.Tracer
}

// Construct Redis client
func New(logger *log.Logger, tracer trace.Tracer) *ImageCache {

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisAddress := fmt.Sprintf("%s:%s", redisHost, redisPort)
	logger.Printf(redisHost)
	logger.Printf(redisPort)
	logger.Printf(redisAddress)

	client := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	return &ImageCache{
		cli:    client,
		logger: logger,
		Tracer: tracer,
	}
}

func (pc *ImageCache) Ping() {
	val, _ := pc.cli.Ping().Result()
	pc.logger.Println(val)
}

//func (ic *ImageCache) PostImage(imageID string, accID string, imageData []byte) error {
//	key := constructImageKey(imageID, accID)
//
//	encodedImage := base64.StdEncoding.EncodeToString(imageData)
//
//	err := ic.cli.Set(key, encodedImage, 300*time.Second).Err()
//	if err != nil {
//		fmt.Println("Error setting image in Redis:", err)
//		return err
//	}
//	return err
//}

func (ic *ImageCache) PostImage(imageID string, accID string, imageData []byte, ctx context.Context) error {
	ctx, span := ic.Tracer.Start(ctx, "ImageCache.PostImage")
	defer span.End()

	key := constructImageKey(imageID, accID)

	encodedImage := base64.StdEncoding.EncodeToString(imageData)

	err := ic.cli.Set(key, encodedImage, 300*time.Second).Err()
	if err != nil {
		span.SetStatus(codes.Error, "Error setting image in Redis"+err.Error())
		fmt.Println("Error setting image in Redis:", err)
		return err
	}
	return nil
}

func (ic *ImageCache) GetImage(imageID, accID string, ctx context.Context) ([]byte, error) {
	ctx, span := ic.Tracer.Start(ctx, "ImageCache.GetImage")
	defer span.End()

	key := constructImageKey(imageID, accID)
	imageData, err := ic.cli.Get(key).Bytes()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	ic.logger.Println("Image Cache hit")
	return imageData, nil
}

func (ic *ImageCache) ImageExists(imageID, accID string, ctx context.Context) bool {
	ctx, span := ic.Tracer.Start(ctx, "ImageCache.ImageExists")
	defer span.End()

	key := constructImageKey(imageID, accID)
	cnt, err := ic.cli.Exists(key).Result()
	if cnt == 1 {
		return true
	}
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false
	}
	return false
}

func (ic *ImageCache) CacheImage(imageID, accID string, imageData string, ctx context.Context) error {
	ctx, span := ic.Tracer.Start(ctx, "ImageCache.CacheImage")
	defer span.End()

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

func (ic *ImageCache) GetAccommodationImages(accommodationID string, ctx context.Context) ([]string, error) {
	ctx, span := ic.Tracer.Start(ctx, "ImageCache.GetAccommodationImages")
	defer span.End()

	cacheKey := fmt.Sprintf(cacheAll, accommodationID)

	images, err := ic.cli.LRange(cacheKey, 0, -1).Result()
	if err == nil {
		span.SetStatus(codes.Error, err.Error())
		return images, nil
	}

	return []string{}, nil
}

func (ic *ImageCache) PostAll(accID string, images []*Image, ctx context.Context) error {
	ctx, span := ic.Tracer.Start(ctx, "ImageCache.PostAll")
	defer span.End()

	cacheKey := fmt.Sprintf(cacheAll, accID)

	for _, image := range images {
		key := constructImageKey(image.ID, accID)
		encodedImage := base64.StdEncoding.EncodeToString(image.Data)

		err := ic.cli.RPush(cacheKey, encodedImage).Err()
		if err != nil {
			span.SetStatus(codes.Error, "Error posting image to Redis:"+err.Error())
			fmt.Println("Error posting image to Redis:", err)
			return err
		}

		err = ic.cli.Set(key, encodedImage, 300*time.Second).Err()
		if err != nil {
			span.SetStatus(codes.Error, "Error setting image in Redis:"+err.Error())
			fmt.Println("Error setting image in Redis:", err)
			return err
		}
	}

	return nil
}
