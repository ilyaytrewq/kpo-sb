package service

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Service) Download(ctx context.Context, bucket, key string) (io.ReadCloser, string, error) {
	out, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, "", err
	}

	ct := ""
	if out.ContentType != nil {
		ct = *out.ContentType
	}
	return out.Body, ct, nil
}
