package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	UploadDir string
	BaseURL   string
}

// NewUploadHandler 创建文件上传处理器
func NewUploadHandler(uploadDir, baseURL string) *UploadHandler {
	return &UploadHandler{
		UploadDir: uploadDir,
		BaseURL:   baseURL,
	}
}

// UploadImage 上传图片
// @Summary      上传图片
// @Description  上传图片文件，支持 jpg/png/gif/bmp/webp/svg 格式，最大 5MB
// @Tags         文件上传
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "图片文件"
// @Success      200  {object}  handler.Response{data=object{url=string}}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /upload/image [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, 400, "请选择要上传的文件")
		return
	}
	defer file.Close()

	// 检查文件大小（最大 5MB）
	if header.Size > 5*1024*1024 {
		Error(c, 400, "文件大小不能超过5MB")
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".webp": true, ".svg": true,
	}
	if !allowedExts[ext] {
		Error(c, 400, "只支持上传图片文件(jpg/png/gif/bmp/webp/svg)")
		return
	}

	// 按日期创建子目录
	dateDir := time.Now().Format("2006/01/02")
	dir := filepath.Join(h.UploadDir, "images", dateDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		Error(c, 500, "创建上传目录失败")
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(dir, filename)

	// 保存文件
	dst, err := os.Create(filePath)
	if err != nil {
		Error(c, 500, "保存文件失败")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		Error(c, 500, "保存文件失败")
		return
	}

	// 返回访问 URL
	url := fmt.Sprintf("%s/images/%s/%s", h.BaseURL, dateDir, filename)

	// wangeditor 要求的返回格式
	c.JSON(http.StatusOK, gin.H{
		"errno": 0,
		"data": gin.H{
			"url":            url,
			"alt":            header.Filename,
			"href":           url,
		},
	})
}
