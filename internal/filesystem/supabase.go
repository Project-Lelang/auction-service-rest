package filesystem

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	storage_go "github.com/supabase-community/storage-go"
)

// supabasePrefixPath is prepended to all object keys inside the bucket.
const supabasePrefixPath = "apis/"

// supabaseStorageHost is the public base URL template for Supabase Storage objects.
const supabaseStorageHost = "%s/storage/v1/object/public"

type supabaseClient struct {
	config *SupabaseClientConfig
}

// SupabaseClientConfig holds the Supabase project credentials and bucket info.
type SupabaseClientConfig struct {
	Client     *storage_go.Client
	ProjectURL string
	ServiceKey string
	BucketName string
}

func (c *supabaseClient) Delete(path string) error {
	_, err := c.config.Client.RemoveFile(c.config.BucketName, []string{supabasePrefixPath + ParsePathTrim(path)})
	return err
}

func (c *supabaseClient) DeleteFolderContents(folderPath string) error {
	prefix := supabasePrefixPath + ParsePathTrim(folderPath)
	files, err := c.config.Client.ListFiles(c.config.BucketName, prefix, storage_go.FileSearchOptions{
		Limit:  1000,
		Offset: 0,
		SortByOptions: storage_go.SortBy{
			Column: "name",
			Order:  "asc",
		},
	})
	if err != nil {
		return fmt.Errorf("error listing objects: %w", err)
	}
	if len(files) == 0 {
		return nil
	}
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = prefix + "/" + f.Name
	}
	_, err = c.config.Client.RemoveFile(c.config.BucketName, paths)
	return err
}

func (c *supabaseClient) Has(path string) (bool, error) {
	data, err := c.config.Client.DownloadFile(c.config.BucketName, supabasePrefixPath+ParsePathTrim(path))
	if err != nil {
		// Supabase returns an error for non-existent objects
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return len(data) >= 0, nil
}

func (c *supabaseClient) Open(_ string) (io.ReadSeekCloser, error) {
	return nil, errors.New("Open not supported on Supabase Storage client")
}

func (c *supabaseClient) Stream(ctx context.Context, path string) (ReadCloserWithContent, error) {
	data, err := c.config.Client.DownloadFile(c.config.BucketName, supabasePrefixPath+ParsePathTrim(path))
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrFileNotExist
		}
		return nil, err
	}
	contentType := http.DetectContentType(data)
	stream := newIoStream(int64(len(data)), contentType)
	go func() {
		_, err := io.Copy(stream.pw, bytes.NewReader(data))
		if err != nil {
			stream.err = err
		}
		stream.closeWriter()
	}()
	return stream, nil
}

func (c *supabaseClient) Url(path string) string {
	base := strings.TrimRight(c.config.ProjectURL, "/")
	return fmt.Sprintf("%s/%s/%s%s", fmt.Sprintf(supabaseStorageHost, base), c.config.BucketName, supabasePrefixPath, ParsePathTrim(path))
}

// PresignedUrl returns a Supabase signed download URL.
func (c *supabaseClient) PresignedUrl(_ string, path string, expiry time.Duration) string {
	expiresIn := int(expiry.Seconds())
	res, err := c.config.Client.CreateSignedUrl(c.config.BucketName, supabasePrefixPath+ParsePathTrim(path), expiresIn)
	if err != nil {
		return ""
	}
	return res.SignedURL
}

// VerifyPresignedUrl is a no-op for Supabase: verification is handled by Supabase itself.
func (c *supabaseClient) VerifyPresignedUrl(_ string) (string, error) {
	return "", nil
}

func (c *supabaseClient) CopyTo(_ context.Context, _ string, _ string) error {
	return errors.New("CopyTo not implemented for Supabase Storage client")
}

func (c *supabaseClient) Write(_ context.Context, r io.ReadSeekCloser, path string) error {
	_, err := c.config.Client.UploadFile(c.config.BucketName, supabasePrefixPath+ParsePathTrim(path), r)
	return err
}

// isNotFoundError checks whether the error from supabase-storage-go indicates a 404.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "404") || strings.Contains(msg, "not found") || strings.Contains(msg, "Object not found")
}

// NewSupabaseClient returns a Supabase-backed filesystem Client.
func NewSupabaseClient(config *SupabaseClientConfig) Client {
	return &supabaseClient{config: config}
}
