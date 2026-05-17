package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// gcsPrefixPath is prepended to all object keys inside the bucket.
const gcsPrefixPath = "apis/"

// gcsStorageHost is the public base URL for GCS objects.
const gcsStorageHost = "https://storage.googleapis.com"

type gcsClient struct {
	config *GcsClientConfig
}

// GcsClientConfig holds the GCS service-account credentials and bucket info.
type GcsClientConfig struct {
	Client      *storage.Client
	ClientEmail string
	PrivateKey  string
	ProjectId   string
	BucketName  string
}

func (c *gcsClient) Delete(path string) error {
	return c.config.Client.Bucket(c.config.BucketName).
		Object(gcsPrefixPath + ParsePathTrim(path)).
		Delete(context.Background())
}

func (c *gcsClient) DeleteFolderContents(folderPath string) error {
	ctx := context.Background()
	it := c.config.Client.Bucket(c.config.BucketName).Objects(ctx, &storage.Query{
		Prefix: gcsPrefixPath + ParsePathTrim(folderPath),
	})
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return fmt.Errorf("error listing objects: %w", err)
		}
		if err := c.config.Client.Bucket(c.config.BucketName).Object(attrs.Name).Delete(ctx); err != nil {
			return fmt.Errorf("error deleting object %s: %w", attrs.Name, err)
		}
	}
	return nil
}

func (c *gcsClient) Has(path string) (bool, error) {
	_, err := c.config.Client.Bucket(c.config.BucketName).
		Object(gcsPrefixPath + ParsePathTrim(path)).
		Attrs(context.Background())
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *gcsClient) Open(_ string) (io.ReadSeekCloser, error) {
	return nil, errors.New("Open not supported on GCS client")
}

func (c *gcsClient) Stream(ctx context.Context, path string) (ReadCloserWithContent, error) {
	obj := c.config.Client.Bucket(c.config.BucketName).Object(gcsPrefixPath + ParsePathTrim(path))
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, ErrFileNotExist
		}
		return nil, err
	}

	stream := newIoStream(attrs.Size, attrs.ContentType)
	go func() {
		reader, err := obj.NewReader(ctx)
		if err != nil {
			stream.err = err
			stream.closeWriter()
			return
		}
		defer reader.Close()
		if _, err = io.Copy(stream.pw, reader); err != nil {
			stream.err = err
		}
		stream.closeWriter()
	}()
	return stream, nil
}

func (c *gcsClient) Url(path string) string {
	return fmt.Sprintf("%s/%s/%s%s", gcsStorageHost, c.config.BucketName, gcsPrefixPath, ParsePathTrim(path))
}

// PresignedUrl returns a GCS V4 signed download URL.
func (c *gcsClient) PresignedUrl(_ string, path string, expiry time.Duration) string {
	opts := &storage.SignedURLOptions{
		Method:         "GET",
		Expires:        time.Now().Add(expiry),
		Scheme:         storage.SigningSchemeV4,
		GoogleAccessID: c.config.ClientEmail,
		PrivateKey:     []byte(c.config.PrivateKey),
	}
	signed, err := storage.SignedURL(c.config.BucketName, gcsPrefixPath+ParsePathTrim(path), opts)
	if err != nil {
		return ""
	}
	return signed
}

// VerifyPresignedUrl is a no-op for GCS: verification is handled by GCS itself.
func (c *gcsClient) VerifyPresignedUrl(_ string) (string, error) {
	return "", nil
}

func (c *gcsClient) CopyTo(_ context.Context, _ string, _ string) error {
	return errors.New("CopyTo not implemented for GCS client")
}

func (c *gcsClient) Write(ctx context.Context, r io.ReadSeekCloser, path string) error {
	wc := c.config.Client.Bucket(c.config.BucketName).
		Object(gcsPrefixPath + ParsePathTrim(path)).
		NewWriter(ctx)
	defer wc.Close()
	_, err := io.Copy(wc, r)
	return err
}

// NewGcsClient returns a GCS-backed filesystem Client.
func NewGcsClient(config *GcsClientConfig) Client {
	return &gcsClient{config: config}
}
