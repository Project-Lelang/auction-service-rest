package filesystem

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	LocalMainPath = "public"
	LocalTmpPath  = "tmp"
)

type localClient struct {
	config    *LocalClientConfig
	chunkSize int
}

// LocalClientConfig holds the local filesystem driver configuration.
type LocalClientConfig struct {
	BasePath              string
	BaseUrl               string
	PresignedUrlSecretKey string
}

func (l *localClient) fullpath(path string) (string, error) {
	path = ParsePathTrim(path)

	if strings.HasPrefix(path, "./") {
		return "", ErrPathIsNotAllowed
	}

	if regexp.MustCompile(`(\.\.\/)+`).MatchString(path) {
		return "", ErrPathIsNotAllowed
	}

	if strings.HasSuffix(path, "/..") {
		return "", ErrPathIsNotAllowed
	}

	return fmt.Sprintf("%s/%s", l.config.BasePath, path), nil
}

func (l *localClient) Delete(path string) error {
	fp, err := l.fullpath(path)
	if err != nil {
		return err
	}
	return os.Remove(fp)
}

func (l *localClient) DeleteFolderContents(folderPath string) error {
	fp, err := l.fullpath(folderPath)
	if err != nil {
		return err
	}
	fi, err := os.Stat(fp)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return ErrPathIsNotDirectory
	}
	return os.RemoveAll(fp)
}

func (l *localClient) Has(path string) (bool, error) {
	fp, err := l.fullpath(path)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(fp); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *localClient) Open(path string) (io.ReadSeekCloser, error) {
	fp, err := l.fullpath(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotExist
		}
		return nil, err
	}
	return f, nil
}

func (l *localClient) Stream(ctx context.Context, path string) (ReadCloserWithContent, error) {
	fp, err := l.fullpath(path)
	if err != nil {
		return nil, err
	}
	return newFileStream(fp)
}

func (l *localClient) Url(path string) string {
	fmt.Println("BASEURL =", l.config.BaseUrl)
	return fmt.Sprintf("%s/%s", l.config.BaseUrl, ParsePathTrim(path))
}

// PresignedUrl generates an HMAC-signed time-limited URL for local file serving.
func (l *localClient) PresignedUrl(filename string, path string, expiry time.Duration) string {
	fmt.Println("PATH =", path)
	fmt.Println("URL  =", l.Url(path))
	expires := time.Now().Add(expiry).Unix()
	data := fmt.Sprintf("%s:%d", filename, expires)

	mac := hmac.New(sha256.New, []byte(l.config.PresignedUrlSecretKey))
	mac.Write([]byte(data))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	v := url.Values{}
	v.Set("file", filename)
	v.Set("expires", strconv.FormatInt(expires, 10))
	v.Set("sig", sig)

	return fmt.Sprintf("%s?%s", l.Url(path), v.Encode())
}

// VerifyPresignedUrl parses and validates the HMAC signature.
// Returns the filename on success.
func (l *localClient) VerifyPresignedUrl(rawUrl string) (string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	q := parsedUrl.Query()
	file := q.Get("file")
	expiresStr := q.Get("expires")
	sig := q.Get("sig")

	if file == "" || expiresStr == "" || sig == "" {
		return "", errors.New("missing required query parameters")
	}

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid expiration timestamp: %w", err)
	}

	if time.Now().Unix() > expires {
		return "", errors.New("URL has expired")
	}

	data := fmt.Sprintf("%s:%d", file, expires)
	mac := hmac.New(sha256.New, []byte(l.config.PresignedUrlSecretKey))
	mac.Write([]byte(data))
	expected := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", errors.New("invalid signature")
	}

	return file, nil
}

func (l *localClient) CopyTo(ctx context.Context, fromPath string, toPath string) error {
	f, err := l.Open(fromPath)
	if err != nil {
		return err
	}
	return l.Write(ctx, f, toPath)
}

func (l *localClient) Write(ctx context.Context, r io.ReadSeekCloser, path string) error {
	fp, err := l.fullpath(path)
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(fp), os.ModePerm)

	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	chunk := make([]byte, l.chunkSize)
	for {
		select {
		case <-ctx.Done():
			l.Delete(path)
			return ctx.Err()
		default:
			n, err := r.Read(chunk)
			if n > 0 {
				if _, writeErr := f.Write(chunk[:n]); writeErr != nil {
					l.Delete(path)
					return writeErr
				}
			}
			if err == io.EOF {
				return nil
			}
			if err != nil {
				l.Delete(path)
				return err
			}
		}
	}
}

// NewLocal returns a local filesystem Client.
func NewLocal(config *LocalClientConfig) Client {
	return &localClient{config: config, chunkSize: defaultChunkSize}
}
