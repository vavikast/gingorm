package controllers

import "mime/multipart"

//文件上传接口
type Uploader interface {
	upload(file multipart.File, fileHeader *multipart.FileHeader) (string, error)
}
