package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// PeerHandler 同行处理器
type PeerHandler struct {
	peerService *service.PeerService
}

// NewPeerHandler 创建同行处理器实例
func NewPeerHandler(peerService *service.PeerService) *PeerHandler {
	return &PeerHandler{peerService: peerService}
}

// Create 创建同行
// POST /api/peers
func (h *PeerHandler) Create(c *gin.Context) {
	var req service.CreatePeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	// 从当前登录用户获取 store_id
	userID := GetUserID(c)
	if userID == 0 {
		Error(c, 401, "未登录")
		return
	}

	peer, err := h.peerService.Create(&req, userID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建同行失败")
		return
	}

	Success(c, peer)
}

// List 获取同行列表
// GET /api/peers?store_id=1&keyword=&page=1&page_size=10
func (h *PeerHandler) List(c *gin.Context) {
	var req service.ListPeerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	result, err := h.peerService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询同行列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// GetDetail 获取同行详情
// GET /api/peers/:id
func (h *PeerHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的同行ID")
		return
	}

	peer, err := h.peerService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取同行详情失败")
		return
	}

	Success(c, peer)
}

// Update 更新同行
// PUT /api/peers/:id
func (h *PeerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的同行ID")
		return
	}

	var req service.UpdatePeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.peerService.Update(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新同行失败")
		return
	}

	Success(c, nil)
}

// Delete 删除同行
// DELETE /api/peers/:id
func (h *PeerHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的同行ID")
		return
	}

	if err := h.peerService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除同行失败")
		return
	}

	Success(c, nil)
}
