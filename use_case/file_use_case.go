package use_case

import (
	"context"
	"fmt"
	"mime/multipart"
	"path"
	"time"

	"auction-service/constant"
	"auction-service/delivery/dto_response"
	internalFilesystem "auction-service/internal/filesystem"
	"auction-service/util"
)

// BaseFileUseCase provides helpers for uploading files to temporary and main storage.
type BaseFileUseCase interface {
	MustValidateFileExtension(ext string, supportedExtensions []string)
	MustValidateFileSize(fileSize int64, maxFileSizeInBytes *int64)
	MustValidateTemporaryFilePaths(paths []string)
	MustCopyFromTmpToMain(ctx context.Context, tmpPath string, mainPath string)
	MustUploadFileToTemporary(ctx context.Context, entityName string, fileHeader *multipart.FileHeader, opt FileUploadTemporaryParams) string
	MustUploadFileFromTemporaryToMain(ctx context.Context, entityName string, entityId int64, destFilename string, tmpPath string, opt FileUploadTemporaryToMainParams) (mainPath string, filename string)
	GetMainFilesystem() internalFilesystem.Client
	GetTmpFilesystem() internalFilesystem.Client
	PresignedLink(path string, expiry time.Duration) string
}

type FileUploadTemporaryParams struct {
	SupportedExtensions []string
	MaxFileSizeInBytes  *int64
}

type FileUploadTemporaryToMainParams struct {
	DeleteTmpOnSuccess bool
}

type baseFileUseCase struct {
	mainFilesystem internalFilesystem.Client
	tmpFilesystem  internalFilesystem.Client
}

// NewBaseFileUseCase creates a BaseFileUseCase backed by the given filesystem clients.
func NewBaseFileUseCase(
	mainFilesystem internalFilesystem.Client,
	tmpFilesystem internalFilesystem.Client,
) BaseFileUseCase {
	return &baseFileUseCase{
		mainFilesystem: mainFilesystem,
		tmpFilesystem:  tmpFilesystem,
	}
}

func (u *baseFileUseCase) MustValidateFileExtension(ext string, supported []string) {
	if supported == nil {
		return
	}
	if !util.StringInSlice(ext, supported) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageFileExtensionIsNotSupported))
	}
}

func (u *baseFileUseCase) MustValidateFileSize(fileSize int64, maxFileSizeInBytes *int64) {
	if fileSize == 0 {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageFileSizeIs0B))
	}
	if maxFileSizeInBytes == nil || fileSize <= *maxFileSizeInBytes {
		return
	}
	maxMB := float64(*maxFileSizeInBytes) / (1024 * 1024)
	panic(dto_response.NewBadRequestErrorResponse(
		constant.LanguageFileMaximumFileSizeIsXMB,
		fmt.Sprintf("%.0f", maxMB),
	))
}

func (u *baseFileUseCase) MustValidateTemporaryFilePaths(paths []string) {
	for _, p := range paths {
		exists, err := u.tmpFilesystem.Has(p)
		panicIfErr(err)
		if !exists {
			if len(paths) > 1 {
				panic(dto_response.NewBadRequestErrorResponse(constant.LanguageFileSomeFileNotExist))
			}
			panic(dto_response.NewBadRequestErrorResponse(constant.LanguageFileFileNotExist))
		}
	}
}

// MustCopyFromTmpToMain copies a file from temporary to main storage.
func (u *baseFileUseCase) MustCopyFromTmpToMain(ctx context.Context, tmpPath string, mainPath string) {
	r, err := u.tmpFilesystem.Open(tmpPath)
	panicIfErr(err)
	defer r.Close()
	panicIfErr(u.mainFilesystem.Write(ctx, r, mainPath))
	go u.tmpFilesystem.Delete(tmpPath)
}

// MustUploadFileToTemporary uploads a multipart file to temporary storage and returns its path.
// The path has the form "{entityName}/{uuid}{ext}".
func (u *baseFileUseCase) MustUploadFileToTemporary(ctx context.Context, entityName string, fileHeader *multipart.FileHeader, opt FileUploadTemporaryParams) string {
	ext := path.Ext(fileHeader.Filename)
	u.MustValidateFileExtension(ext, opt.SupportedExtensions)
	u.MustValidateFileSize(fileHeader.Size, opt.MaxFileSizeInBytes)

	tmpPath := fmt.Sprintf("%s/%s%s", entityName, util.NewUuid(), ext)

	f, err := fileHeader.Open()
	panicIfErr(err)
	defer f.Close()

	panicIfErr(u.tmpFilesystem.Write(ctx, f, tmpPath))
	return tmpPath
}

// MustUploadFileFromTemporaryToMain moves a file from temporary to permanent storage.
// Returns (mainPath, filename).
func (u *baseFileUseCase) MustUploadFileFromTemporaryToMain(ctx context.Context, entityName string, entityId int64, destFilename string, tmpPath string, opt FileUploadTemporaryToMainParams) (string, string) {
	u.MustValidateTemporaryFilePaths([]string{tmpPath})

	mainPath := fmt.Sprintf("%s/%d/%s", entityName, entityId, destFilename)

	r, err := u.tmpFilesystem.Open(tmpPath)
	panicIfErr(err)
	defer r.Close()

	panicIfErr(u.mainFilesystem.Write(ctx, r, mainPath))

	if opt.DeleteTmpOnSuccess {
		go u.tmpFilesystem.Delete(tmpPath)
	}

	return mainPath, util.GetFilenameFromPath(tmpPath)
}

func (u *baseFileUseCase) GetMainFilesystem() internalFilesystem.Client {
	return u.mainFilesystem
}

func (u *baseFileUseCase) GetTmpFilesystem() internalFilesystem.Client {
	return u.tmpFilesystem
}

// PresignedLink generates a time-limited presigned URL for main storage.
func (u *baseFileUseCase) PresignedLink(filePath string, expiry time.Duration) string {
	return u.mainFilesystem.PresignedUrl(util.GetFilenameFromPath(filePath), filePath, expiry)
}
