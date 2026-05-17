package dto_request

import "mime/multipart"

type FileUploadRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}
