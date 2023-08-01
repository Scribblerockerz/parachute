package s3

import (
	"context"
	"errors"
	"strings"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	minioClient *minio.Client
}

func NewClient(endpoint string, accessKey string, secretKey string, useSSL bool) (*S3Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		return nil, err
	}

	return &S3Client{minioClient}, nil
}

type PayloadInfo struct {
	Bucket      string
	Object      string
	FilePath    string
	ContentType string
}

func NewPayload(remote string, filePath string) (*PayloadInfo, error) {
	if !strings.HasPrefix(remote, "s3://") {
		return nil, errors.New("invalid remote target provided. Expected 's3://bucket/object' format")
	}

	remote, _ = strings.CutPrefix(remote, "s3://")
	parts := strings.SplitAfterN(remote, "/", 2)

	return &PayloadInfo{
		Bucket:      strings.Trim(parts[0], "/"),
		Object:      strings.Trim(parts[1], "/"),
		FilePath:    filePath,
		ContentType: "application/octet-stream",
	}, nil
}

type DownloadInfo struct {
	Bucket   string
	Object   string
	FilePath string
}

func NewDownload(remote string, filePath string) (*DownloadInfo, error) {
	if !strings.HasPrefix(remote, "s3://") {
		return nil, errors.New("invalid remote target provided. Expected 's3://bucket/object' format")
	}

	remote, _ = strings.CutPrefix(remote, "s3://")
	parts := strings.SplitAfterN(remote, "/", 2)

	return &DownloadInfo{
		Bucket:   strings.Trim(parts[0], "/"),
		Object:   strings.Trim(parts[1], "/"),
		FilePath: filePath,
	}, nil
}

func (s3 *S3Client) UploadPayload(ctx context.Context, payload *PayloadInfo) (minio.UploadInfo, error) {
	info, err := s3.minioClient.FPutObject(ctx, payload.Bucket, payload.Object, payload.FilePath, minio.PutObjectOptions{
		ContentType: payload.ContentType,
	})
	if err != nil {
		return minio.UploadInfo{}, err
	}

	return info, nil
}

func (s3 *S3Client) DownloadPayload(ctx context.Context, info *DownloadInfo) error {
	return s3.minioClient.FGetObject(ctx, info.Bucket, info.Object, info.FilePath, minio.GetObjectOptions{})
}
