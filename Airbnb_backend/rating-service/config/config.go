package config

import "os"

type Config struct {
	JaegerAddress                     string
	ServiceName                       string
	NatsHost                          string
	NatsPort                          string
	NatsUser                          string
	NatsPass                          string
	CreateAccommodationCommandSubject string
	CreateAccommodationReplySubject   string
}

func GetConfig() Config {
	return Config{
		JaegerAddress:                     os.Getenv("JAEGER_ADDRESS"),
		ServiceName:                       "recommendation-service",
		NatsHost:                          os.Getenv("NATS_HOST"),
		NatsPort:                          os.Getenv("NATS_PORT"),
		NatsUser:                          os.Getenv("NATS_USER"),
		NatsPass:                          os.Getenv("NATS_PASS"),
		CreateAccommodationCommandSubject: os.Getenv("CREATE_ACCOMMODATION_COMMAND_SUBJECT"),
		CreateAccommodationReplySubject:   os.Getenv("CREATE_ACCOMMODATION_REPLY_SUBJECT"),
	}
}
