package config

import "os"

type Config struct {
	JaegerAddress string
	ServiceName   string
}

func GetConfig() Config {
	return Config{
		JaegerAddress: os.Getenv("JAEGER_ADDRESS"),
		ServiceName:   "reservations-service",
	}
}
