package filesystem

import (
	"fmt"
	"strings"

	"auction-service/global"
)

const (
	FilesystemLocal = "local"
	FilesystemGCS   = "gcs"
)

// FilesystemManager exposes main (permanent) and tmp (temporary) filesystem clients.
type FilesystemManager interface {
	Main() Client
	Tmp() Client
}

type filesystemManager struct {
	main Client
	tmp  Client
}

func (m *filesystemManager) Main() Client { return m.main }
func (m *filesystemManager) Tmp() Client  { return m.tmp }

// Config holds optional cloud provider config passed from the application manager.
type Config struct {
	Filesystem      string
	GcsClientConfig *GcsClientConfig
}

func generateLocalClientConfig(subDir string) *LocalClientConfig {
	return &LocalClientConfig{
		BasePath:              fmt.Sprintf("%s/%s", global.GetStorageDir(), subDir),
		BaseUrl:               fmt.Sprintf("%s/%s", strings.TrimRight(global.GetFilesystem().BaseUri, "/"), subDir),
		PresignedUrlSecretKey: global.GetFilesystem().PresignedUrlSecretKey,
	}
}

// NewFilesystemManager creates main + tmp filesystem clients based on config.
// tmp is always local; main follows the configured type.
func NewFilesystemManager(config Config) FilesystemManager {
	var main Client
	switch config.Filesystem {
	case FilesystemGCS:
		main = NewGcsClient(config.GcsClientConfig)
	default:
		main = NewLocal(generateLocalClientConfig(LocalMainPath))
	}

	return &filesystemManager{
		main: main,
		tmp:  NewLocal(generateLocalClientConfig(LocalTmpPath)),
	}
}
