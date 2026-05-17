package api

import (
	"net/http"
	"strings"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	internalFilesystem "auction-service/internal/filesystem"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

type FileApi struct {
	*api
	baseFileUseCase   use_case.BaseFileUseCase
	filesystemManager internalFilesystem.FilesystemManager
}

// UploadTemporary godoc
//
//	@Router		/own/files/upload [post]
//	@Summary	Upload a file to temporary storage and return its temporary path
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		mpfd
//	@Param		file	formData	file	true	"File to upload"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{path=string}}
func (a *FileApi) UploadTemporary() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.FileUploadRequest
		ctx.mustBindForm(&request)

		tmpPath := a.baseFileUseCase.MustUploadFileToTemporary(ctx.context(), "upload", request.File, use_case.FileUploadTemporaryParams{})

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"path": tmpPath,
			},
		})
	})
}

// ServeLocal godoc
//
//	@Router		/storage/public/{filePath} [get]
//	@Summary	Serve a presigned local file
//	@tags		File
//	@Param		filePath	path	string	true	"File path"
//	@Produce	octet-stream
//	@Success	200
func (a *FileApi) ServeLocal() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		rawPath := ctx.getParam("filePath")
		rawPath = strings.TrimPrefix(rawPath, "/")

		rawUrl := ctx.ginCtx.Request.URL.String()
		if _, err := a.filesystemManager.Main().VerifyPresignedUrl(rawUrl); err != nil {
			ctx.json(http.StatusForbidden, dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
			return
		}

		stream, err := a.filesystemManager.Main().Stream(ctx.context(), rawPath)
		if err != nil {
			ctx.json(http.StatusNotFound, dto_response.NewNotFoundErrorResponse(constant.LanguageSystemNotFound))
			return
		}
		defer stream.Close()

		ctx.ginCtx.DataFromReader(http.StatusOK, stream.ContentLength(), stream.ContentType(), stream, nil)
	})
}

func RegisterFileApi(router gin.IRouter, api *api, filesystemManager internalFilesystem.FilesystemManager, baseFileUseCase use_case.BaseFileUseCase) {
	fileApi := FileApi{
		api:               api,
		baseFileUseCase:   baseFileUseCase,
		filesystemManager: filesystemManager,
	}

	own := router.Group("/own")
	own.POST("/files/upload", fileApi.UploadTemporary())

	// local presigned file serving
	router.GET("/storage/public/*filePath", fileApi.ServeLocal())
}
