package database

import (
	"context"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var GarageClient *minio.Client

func OpenGarage() error {
	endpoint := os.Getenv("GARAGE_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:3900"
	}
	accessKey := os.Getenv("GARAGE_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	secretKey := os.Getenv("GARAGE_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadminminioadmin"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
		Region: "garage",
	})
	if err != nil {
		return err
	}
	GarageClient = client

	// teste de conexão
	ctx := context.Background()
	bucket := os.Getenv("GARAGE_BUCKET")
	if bucket == "" {
		bucket = "spotcin-audio"
	}
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		return client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "garage"})
	}
	return nil
}