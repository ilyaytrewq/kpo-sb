package config

import (
	"fmt"
	"os"
)

type S3Config struct {
	Endpoint  string // "https://storage.yandexcloud.net"
	Region    string // "ru-central1"
	AccessKey string // YC static key id
	SecretKey string // YC static secret key
	Bucket   string // "kpo-hw3"
}

func LoadS3ConfigFromEnv() (*S3Config, error) {
	s3_endpoint := os.Getenv("S3_ENDPOINT")
	s3_region := os.Getenv("S3_REGION")
	s3_access_key := os.Getenv("AWS_ACCESS_KEY_ID")
	s3_secret_key := os.Getenv("AWS_SECRET_ACCESS_KEY")
	s3_bucket := os.Getenv("S3_BUCKET")

	if s3_endpoint == "" {
		return nil, fmt.Errorf("S3 endpoint is empty")
	}
	if s3_region == "" {
		return nil, fmt.Errorf("S3 region is empty")
	}
	if s3_access_key == "" {
		return nil, fmt.Errorf("S3 access key is empty")
	}
	if s3_secret_key == "" {
		return nil, fmt.Errorf("S3 secret key is empty")
	}
	if s3_bucket == "" {
		return nil, fmt.Errorf("S3 bucket is empty")
	}
	
	return &S3Config{
		Endpoint:  s3_endpoint,
		Region:    s3_region,
		AccessKey: s3_access_key,
		SecretKey: s3_secret_key,
		Bucket:   s3_bucket,
	}, nil
}