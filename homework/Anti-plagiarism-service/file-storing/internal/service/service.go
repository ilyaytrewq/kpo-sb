package service

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	internal_config "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/config"
)

type Config struct {
	Endpoint  string // "https://storage.yandexcloud.net"
	Region    string // "ru-central1"
	AccessKey string // YC static key id
	SecretKey string // YC static secret key
	Bucket   string // "kpo-hw3"
}

type S3Service struct {
	Client *s3.Client
	Config Config
}

func (s *S3Service) Bucket() string {
	if s == nil {
		return ""
	}
	return s.Config.Bucket
}

func NewService(ctx context.Context) (*S3Service, error) {
	internalConfig, err := internal_config.LoadS3ConfigFromEnv()
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(internalConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(internalConfig.AccessKey, internalConfig.SecretKey, "")),
		
	)
	if err != nil {
		return nil, err
	}

	cli := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(internalConfig.Endpoint)
		o.UsePathStyle = true
	})

	return &S3Service{Client: cli, Config: Config{
		Endpoint:  internalConfig.Endpoint,
		Region:    internalConfig.Region,
		AccessKey: internalConfig.AccessKey,
		SecretKey: internalConfig.SecretKey,
		Bucket: internalConfig.Bucket,
	}}, nil
}
