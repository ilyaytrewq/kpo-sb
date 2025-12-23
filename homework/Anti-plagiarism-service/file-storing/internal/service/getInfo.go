package service

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

/*
	type FileInfoResponse struct {
	    ChecksumSha256   *string            `json:"checksumSha256,omitempty"`
	    ContentType      *string            `json:"contentType,omitempty"`
	    FileId           openapi_types.UUID `json:"fileId"`
	    OriginalFileName *string            `json:"originalFileName,omitempty"`
	    SizeBytes        *int64             `json:"sizeBytes,omitempty"`
	    StoredAt         time.Time          `json:"storedAt"`
	}
*/
type Info struct {
	Size           int64
	ContentType    string
	StoredAt       time.Time
	FileName       string
	ChecksumSha256 string
}

func (s *S3Service) Head(ctx context.Context, bucket, key string) (Info, error) {
	out, err := s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return Info{}, err
	}

	var inf Info
	if out.ContentLength != nil {
		inf.Size = *out.ContentLength
	}
	if out.ChecksumSHA256 != nil {
		inf.ChecksumSha256 = *out.ChecksumSHA256
	}
	if out.ContentType != nil {
		inf.ContentType = *out.ContentType
	}
	if out.LastModified != nil {
		inf.StoredAt = *out.LastModified
	}
	if out.Metadata != nil {
		if filename, ok := out.Metadata["original-file-name"]; ok {
			inf.FileName = filename
		}
	}
	return inf, nil
}
