package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code      int         `json:"code" example:"200"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp" example:"1715000000"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total" example:"100"`
	Page     int         `json:"page" example:"1"`
	PageSize int         `json:"page_size" example:"10"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      200,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      200,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Unix(),
	})
}

// ErrorWithHTTPStatus 错误响应（自定义HTTP状态码）
func ErrorWithHTTPStatus(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Unix(),
	})
}

// PageResponse 分页响应
func PageResponse(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		Timestamp: time.Now().Unix(),
	})
}
