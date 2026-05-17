package filesystem

import (
	"context"
	"io"
	"os"
	"time"
)

const defaultChunkSize = 5 << 20 // 5 MB

// Client abstracts access to any underlying object storage provider.
type Client interface {
	Delete(path string) error
	DeleteFolderContents(folderPath string) error
	Has(path string) (bool, error)
	Open(path string) (io.ReadSeekCloser, error)
	Stream(ctx context.Context, path string) (ReadCloserWithContent, error)
	Url(path string) string
	PresignedUrl(filename string, path string, expiry time.Duration) string
	VerifyPresignedUrl(rawUrl string) (string, error)
	CopyTo(ctx context.Context, fromPath string, toPath string) error
	Write(ctx context.Context, r io.ReadSeekCloser, path string) error
}

// ReadCloserWithContent adds metadata to the standard ReadCloser.
type ReadCloserWithContent interface {
	ContentLength() int64
	ContentType() string
	io.ReadCloser
}

// ioStream is a pipe-backed ReadCloserWithContent used for cloud providers.
type ioStream struct {
	contentLength int64
	contentType   string
	pr            *io.PipeReader
	pw            *io.PipeWriter
	err           error
}

func (s *ioStream) closeWriter() error { return s.pw.Close() }

func (s *ioStream) ContentLength() int64 { return s.contentLength }
func (s *ioStream) ContentType() string  { return s.contentType }

func (s *ioStream) Read(p []byte) (n int, err error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.pr.Read(p)
}

func (s *ioStream) Close() error { return s.pr.Close() }

// WriteAt is used by the S3 downloader (sequential only).
func (s *ioStream) WriteAt(p []byte, _ int64) (n int, err error) {
	return s.pw.Write(p)
}

func newIoStream(contentLength int64, contentType string) *ioStream {
	pr, pw := io.Pipe()
	return &ioStream{contentLength: contentLength, contentType: contentType, pr: pr, pw: pw}
}

// fileStream wraps an *os.File as ReadCloserWithContent.
type fileStream struct {
	contentLength int64
	contentType   string
	file          *os.File
}

func (f *fileStream) ContentLength() int64       { return f.contentLength }
func (f *fileStream) ContentType() string        { return f.contentType }
func (f *fileStream) Read(p []byte) (int, error) { return f.file.Read(p) }
func (f *fileStream) Close() error               { return f.file.Close() }

func newFileStream(path string) (*fileStream, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotExist
		}
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	ct, err := guessMimeType(f)
	if err != nil {
		return nil, err
	}

	return &fileStream{contentLength: fi.Size(), contentType: ct, file: f}, nil
}
