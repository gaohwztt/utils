package cos

import (
	"mime/multipart"
	"path"
	"strings"
)

const (
	// DurationSeconds
	durationSecondsDefault = 3600
	durationSecondsMax     = 7200

	// Effect
	effectFalse = "deny"
	effectTrue  = "allow"

	// Resource
	resourceDefault = "qcs::cos:%s:uid/%s:%s/*"
)

func actionDefault() []string {
	return []string{
		"name/cos:PostObject",
		"name/cos:PutObject",
		"name/cos:GetObject",
	}
}

// 判断是否是图片
func isImg(file *multipart.FileHeader) string {
	fileExt := strings.ToLower(path.Ext(file.Filename))
	if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".gif" && fileExt != ".jpeg" && fileExt != ".webp" {
		return "上传失败!只允许png,jpg,gif,jpeg,webp文件"
	}
	return ""
}
