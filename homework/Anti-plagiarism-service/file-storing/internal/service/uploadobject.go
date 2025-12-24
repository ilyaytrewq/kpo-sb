package service

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Service) Upload(ctx context.Context, bucket, key, contentType, originalFileName string, body io.Reader, size int64) error {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &bucket,
		Key:           &key,
		Body:          body,
		ContentType:   &contentType,
		ContentLength: &size,
		Metadata: map[string]string{
			"original-file-name": originalFileName,
		},
	}, s3.WithAPIOptions(
		v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
	))
	return err
}
