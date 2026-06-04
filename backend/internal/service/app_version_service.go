package service

import (
	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// CreateAppVersionRequest 创建APP版本请求
type CreateAppVersionRequest struct {
	Platform      string `json:"platform" binding:"required,oneof=ios android"`
	VersionCode   string `json:"version_code" binding:"required"`
	VersionName   string `json:"version_name" binding:"required"`
	DownloadURL   string `json:"download_url" binding:"required,url"`
	FileSize      int64  `json:"file_size"`
	UpdateType    string `json:"update_type" binding:"required,oneof=full wgt"` // 更新类型: full-整包更新, wgt-热更新
	IsForceUpdate int8   `json:"is_force_update"`
	MinVersion    string `json:"min_version"`
	UpdateContent string `json:"update_content"`
	Status        int8   `json:"status"`
}

// UpdateAppVersionRequest 更新APP版本请求
type UpdateAppVersionRequest struct {
	VersionName   string `json:"version_name"`
	DownloadURL   string `json:"download_url"`
	UpdateType    string `json:"update_type"` // 更新类型: full-整包更新, wgt-热更新
	IsForceUpdate int8   `json:"is_force_update"`
	MinVersion    string `json:"min_version"`
	UpdateContent string `json:"update_content"`
	Status        int8   `json:"status"`
}

// ListAppVersionRequest 版本列表查询请求
type ListAppVersionRequest struct {
	Platform string `form:"platform"`
	Status   *int8  `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// AppVersionResponse APP版本响应（update_type转为字符串）
type AppVersionResponse struct {
	ID             int64      `json:"id"`
	Platform       string     `json:"platform"`
	VersionCode    string     `json:"version_code"`
	VersionName    string     `json:"version_name"`
	DownloadURL    string     `json:"download_url"`
	FileSize       int64      `json:"file_size"`
	UpdateType     string     `json:"update_type"` // full-整包更新, wgt-热更新
	IsForceUpdate  int8       `json:"is_force_update"`
	MinVersion     string     `json:"min_version"`
	UpdateContent  string     `json:"update_content"`
	Status         int8       `json:"status"`
	PublishedAt    *time.Time `json:"published_at"`
	CreatedBy      *int64     `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// convertToAppVersionResponse 将模型转换为响应（update_type转字符串）
func convertToAppVersionResponse(v *models.AppVersion) *AppVersionResponse {
	if v == nil {
		return nil
	}
	updateType := "full"
	if v.UpdateType == 1 {
		updateType = "wgt"
	}
	return &AppVersionResponse{
		ID:             v.ID,
		Platform:       v.Platform,
		VersionCode:    v.VersionCode,
		VersionName:    v.VersionName,
		DownloadURL:    v.DownloadURL,
		FileSize:       v.FileSize,
		UpdateType:     updateType,
		IsForceUpdate:  v.IsForceUpdate,
		MinVersion:     v.MinVersion,
		UpdateContent:  v.UpdateContent,
		Status:         v.Status,
		PublishedAt:    v.PublishedAt,
		CreatedBy:      v.CreatedBy,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

// AppVersionService APP版本服务
type AppVersionService struct {
	db  *gorm.DB
	repo *repository.BaseRepository[models.AppVersion]
}

// NewAppVersionService 创建APP版本服务实例
func NewAppVersionService(db *gorm.DB) *AppVersionService {
	return &AppVersionService{
		db:   db,
		repo: repository.NewBaseRepository[models.AppVersion](db),
	}
}

// Create 创建APP版本
func (s *AppVersionService) Create(req *CreateAppVersionRequest, createdBy int64) (*models.AppVersion, error) {
	// 检查版本号是否已存在
	var existing models.AppVersion
	if err := s.db.Where("platform = ? AND version_code = ?", req.Platform, req.VersionCode).First(&existing).Error; err == nil {
		return nil, &AppError{Code: 400, Message: "该平台的此版本号已存在"}
	}

	// 文件大小自动读取
	fileSize := req.FileSize
	if fileSize == 0 && req.DownloadURL != "" {
		size, err := getFileSizeFromURL(req.DownloadURL)
		if err == nil {
			fileSize = size
		}
	}

	// 转换更新类型: full-整包更新(0), wgt-热更新(1)
	var updateType int8
	if req.UpdateType == "wgt" {
		updateType = 1
	} else {
		updateType = 0 // 默认整包更新
	}

	version := &models.AppVersion{
		Platform:      req.Platform,
		VersionCode:   req.VersionCode,
		VersionName:   req.VersionName,
		DownloadURL:   req.DownloadURL,
		FileSize:      fileSize,
		UpdateType:    updateType,
		IsForceUpdate: req.IsForceUpdate,
		MinVersion:    req.MinVersion,
		UpdateContent: req.UpdateContent,
		Status:        req.Status,
		CreatedBy:     &createdBy,
	}

	if err := s.repo.Create(version); err != nil {
		return nil, &AppError{Code: 500, Message: "创建版本失败"}
	}
	return version, nil
}

// Update 更新APP版本
func (s *AppVersionService) Update(id int64, req *UpdateAppVersionRequest) (*models.AppVersion, error) {
	version, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &AppError{Code: 404, Message: "版本不存在"}
		}
		return nil, &AppError{Code: 500, Message: "查询版本失败"}
	}

	updates := map[string]interface{}{}
	if req.VersionName != "" {
		updates["version_name"] = req.VersionName
	}
	if req.DownloadURL != "" {
		updates["download_url"] = req.DownloadURL
		// 如果下载地址变更，尝试自动获取文件大小
		size, err := getFileSizeFromURL(req.DownloadURL)
		if err == nil {
			updates["file_size"] = size
		}
	}
	if req.UpdateType != "" {
		// 转换更新类型: full-整包更新(0), wgt-热更新(1)
		if req.UpdateType == "wgt" {
			updates["update_type"] = 1
		} else {
			updates["update_type"] = 0
		}
	}
	updates["is_force_update"] = req.IsForceUpdate
	updates["min_version"] = req.MinVersion
	updates["update_content"] = req.UpdateContent
	updates["status"] = req.Status

	if err := s.db.Model(version).Updates(updates).Error; err != nil {
		return nil, &AppError{Code: 500, Message: "更新版本失败"}
	}

	return version, nil
}

// Delete 删除APP版本
func (s *AppVersionService) Delete(id int64) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &AppError{Code: 404, Message: "版本不存在"}
		}
		return &AppError{Code: 500, Message: "查询版本失败"}
	}

	if err := s.repo.Delete(id); err != nil {
		return &AppError{Code: 500, Message: "删除版本失败"}
	}
	return nil
}

// GetByID 根据ID获取版本
func (s *AppVersionService) GetByID(id int64) (*AppVersionResponse, error) {
	version, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &AppError{Code: 404, Message: "版本不存在"}
		}
		return nil, &AppError{Code: 500, Message: "查询版本失败"}
	}
	return convertToAppVersionResponse(version), nil
}

// List 获取版本列表
func (s *AppVersionService) List(req *ListAppVersionRequest) (*PageResult, error) {
	db := s.db.Model(&models.AppVersion{})

	if req.Platform != "" {
		db = db.Where("platform = ?", req.Platform)
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, &AppError{Code: 500, Message: "查询版本列表失败"}
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var versions []models.AppVersion
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&versions).Error; err != nil {
		return nil, &AppError{Code: 500, Message: "查询版本列表失败"}
	}

	// 转换为响应格式（update_type转为字符串）
	var list []*AppVersionResponse
	for i := range versions {
		list = append(list, convertToAppVersionResponse(&versions[i]))
	}

	return &PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetLatest 获取最新版本（APP端调用）
func (s *AppVersionService) GetLatest(platform string) (*AppVersionResponse, error) {
	var version models.AppVersion
	if err := s.db.Where("platform = ? AND status = 1", platform).
		Order("version_code DESC").
		First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &AppError{Code: 404, Message: "暂无版本信息"}
		}
		return nil, &AppError{Code: 500, Message: "查询版本失败"}
	}
	return convertToAppVersionResponse(&version), nil
}

// getFileSizeFromURL 通过HTTP HEAD请求获取文件大小
func getFileSizeFromURL(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, nil // 无法获取时返回0，不阻断流程
	}

	return resp.ContentLength, nil
}
