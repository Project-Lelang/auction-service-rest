package infrastructure

import (
	"context"
	"os"

	"auction-service/global"

	"cloud.google.com/go/storage"
)

func NewGcsClient(gcsConfig *global.GcsConfig) *storage.Client {
	if gcsConfig == nil {
		return nil
	}

	if _, err := os.Stat(gcsConfig.ConfigFilepath()); err != nil {
		return nil
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gcsConfig.ConfigFilepath())

	client, err := storage.NewClient(context.Background())
	if err != nil {
		panic(err)
	}

	return client
}
