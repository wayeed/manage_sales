package service

import (
	"compress/gzip"
	"furniture-commission/configs"
	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// CreateBackupRequest 创建备份请求
type CreateBackupRequest struct {
	BackupType int8 `json:"backup_type" binding:"required"` // 0-手动, 1-自动
}

// ListBackupRequest 备份列表查询请求
type ListBackupRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ClearDataRequest 清除数据请求
type ClearDataRequest struct {
	TableCategories []string `json:"table_categories" binding:"required"` // 要清除的数据分类
}

// DataTableInfo 数据表信息
type DataTableInfo struct {
	Category    string   `json:"category"`
	Label       string   `json:"label"`
	Tables      []string `json:"tables"`
	Description string   `json:"description"`
}

// MaintenanceService 平台维护服务
type MaintenanceService struct {
	db       *gorm.DB
	repo     *repository.BaseRepository[models.SystemBackup]
	backupDir string
}

// NewMaintenanceService 创建平台维护服务实例
func NewMaintenanceService(db *gorm.DB) *MaintenanceService {
	// 确保备份目录存在
	backupDir := "./backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Printf("[WARN] 创建备份目录失败: %v\n", err)
	}

	return &MaintenanceService{
		db:        db,
		repo:      repository.NewBaseRepository[models.SystemBackup](db),
		backupDir: backupDir,
	}
}

// GetDataTables 获取可清除的数据表列表
func (s *MaintenanceService) GetDataTables() []DataTableInfo {
	return []DataTableInfo{
		{
			Category:    "orders",
			Label:       "订单数据",
			Tables:      []string{"orders", "order_items", "order_gifts"},
			Description: "订单主表及明细",
		},
		{
			Category:    "payments",
			Label:       "回款数据",
			Tables:      []string{"payments"},
			Description: "回款记录",
		},
		{
			Category:    "inventory",
			Label:       "库存数据",
			Tables:      []string{"warehouse_stocks", "warehouse_gift_stocks", "inventory_transactions"},
			Description: "库存记录及流水",
		},
		{
			Category:    "commissions",
			Label:       "提成数据",
			Tables:      []string{"commissions", "fund_pools", "fund_pool_shares"},
			Description: "提成记录及基金池",
		},
		{
			Category:    "salaries",
			Label:       "工资数据",
			Tables:      []string{"salary_records", "salary_items"},
			Description: "工资记录",
		},
		{
			Category:    "alerts",
			Label:       "预警数据",
			Tables:      []string{"stock_alerts"},
			Description: "库存预警记录",
		},
		{
			Category:    "customers",
			Label:       "客户数据",
			Tables:      []string{"customers", "peers"},
			Description: "客户及同行信息",
		},
	}
}

// CheckRecentBackup 检查是否有10分钟内的备份（状态为进行中或成功即可）
func (s *MaintenanceService) CheckRecentBackup() (bool, *models.SystemBackup, error) {
	var backup models.SystemBackup
	tenMinutesAgo := time.Now().Add(-10 * time.Minute)

	// 只要备份已开始（状态0=进行中或1=成功）即可
	err := s.db.Where("status IN (0, 1) AND created_at >= ?", tenMinutesAgo).
		Order("created_at DESC").
		First(&backup).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	}

	return true, &backup, nil
}

// CreateBackup 创建数据库备份
func (s *MaintenanceService) CreateBackup(req *CreateBackupRequest, createdBy int64) (*models.SystemBackup, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupTypeStr := "manual"
	if req.BackupType == 1 {
		backupTypeStr = "auto"
	}
	fileName := fmt.Sprintf("backup_%s_%s.sql.gz", backupTypeStr, timestamp)
	filePath := filepath.Join(s.backupDir, fileName)

	// 创建备份记录
	backup := &models.SystemBackup{
		BackupType: req.BackupType,
		FileName:   fileName,
		FilePath:   filePath,
		Status:     0, // 进行中
		CreatedBy:  &createdBy,
	}

	if err := s.repo.Create(backup); err != nil {
		return nil, &AppError{Code: 500, Message: "创建备份记录失败"}
	}

	// 异步执行备份
	go s.executeBackup(backup)

	return backup, nil
}

// executeBackup 执行备份命令（gzip压缩）
func (s *MaintenanceService) executeBackup(backup *models.SystemBackup) {
	cfg := configs.GlobalConfig
	if cfg == nil {
		s.updateBackupStatus(backup.ID, 2, "读取系统配置失败")
		return
	}

	dbName := cfg.Database.DBName
	dbUser := cfg.Database.User
	dbPass := cfg.Database.Password
	dbHost := cfg.Database.Host
	dbPort := cfg.Database.Port

	// mysqldump 命令
	dumpArgs := []string{
		fmt.Sprintf("-u%s", dbUser),
		fmt.Sprintf("-p%s", dbPass),
		fmt.Sprintf("-h%s", dbHost),
		fmt.Sprintf("-P%d", dbPort),
		"--single-transaction",
		"--routines",
		"--triggers",
		"--events",
		dbName,
	}

	cmd := exec.Command("mysqldump", dumpArgs...)

	// 创建 gzip 压缩文件
	gzFile, err := os.Create(backup.FilePath)
	if err != nil {
		s.updateBackupStatus(backup.ID, 2, "创建备份文件失败: "+err.Error())
		return
	}
	defer gzFile.Close()

	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()

	cmd.Stdout = gzWriter
	if err := cmd.Run(); err != nil {
		s.updateBackupStatus(backup.ID, 2, "执行mysqldump失败: "+err.Error())
		return
	}

	// 确保 gzip 写入完成
	gzWriter.Flush()

	// 获取文件大小
	fileInfo, err := os.Stat(backup.FilePath)
	if err != nil {
		s.updateBackupStatus(backup.ID, 2, "获取备份文件大小失败: "+err.Error())
		return
	}

	now := time.Now()
	s.db.Model(&models.SystemBackup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
		"status":      1,
		"file_size":   fileInfo.Size(),
		"finished_at": &now,
	})
}

// updateBackupStatus 更新备份状态
func (s *MaintenanceService) updateBackupStatus(id int64, status int8, errorMsg string) {
	now := time.Now()
	s.db.Model(&models.SystemBackup{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"error_message": errorMsg,
		"finished_at":   &now,
	})
}

// ListBackups 获取备份列表
func (s *MaintenanceService) ListBackups(req *ListBackupRequest) (*PageResult, error) {
	db := s.db.Model(&models.SystemBackup{})

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, &AppError{Code: 500, Message: "查询备份列表失败"}
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

	var backups []models.SystemBackup
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&backups).Error; err != nil {
		return nil, &AppError{Code: 500, Message: "查询备份列表失败"}
	}

	return &PageResult{
		List:     backups,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetBackupByID 根据ID获取备份
func (s *MaintenanceService) GetBackupByID(id int64) (*models.SystemBackup, error) {
	backup, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &AppError{Code: 404, Message: "备份记录不存在"}
		}
		return nil, &AppError{Code: 500, Message: "查询备份记录失败"}
	}
	return backup, nil
}

// DeleteBackup 删除备份
func (s *MaintenanceService) DeleteBackup(id int64) error {
	backup, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &AppError{Code: 404, Message: "备份记录不存在"}
		}
		return &AppError{Code: 500, Message: "查询备份记录失败"}
	}

	// 删除文件
	if err := os.Remove(backup.FilePath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("[WARN] 删除备份文件失败: %v\n", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return &AppError{Code: 500, Message: "删除备份记录失败"}
	}
	return nil
}

// RestoreBackup 还原备份
func (s *MaintenanceService) RestoreBackup(id int64) error {
	backup, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &AppError{Code: 404, Message: "备份记录不存在"}
		}
		return &AppError{Code: 500, Message: "查询备份记录失败"}
	}

	if backup.Status != 1 {
		return &AppError{Code: 400, Message: "备份文件不可用"}
	}

	// 检查文件是否存在
	if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
		return &AppError{Code: 404, Message: "备份文件不存在"}
	}

	cfg := configs.GlobalConfig
	if cfg == nil {
		return &AppError{Code: 500, Message: "读取系统配置失败"}
	}

	dbName := cfg.Database.DBName
	dbUser := cfg.Database.User
	dbPass := cfg.Database.Password
	dbHost := cfg.Database.Host
	dbPort := cfg.Database.Port

	// 打开备份文件（支持 gzip 解压）
	inFile, err := os.Open(backup.FilePath)
	if err != nil {
		return &AppError{Code: 500, Message: "打开备份文件失败"}
	}
	defer inFile.Close()

	var reader *os.File = inFile
	// 如果是 gzip 文件则解压
	if filepath.Ext(backup.FilePath) == ".gz" {
		gzReader, err := gzip.NewReader(inFile)
		if err != nil {
			return &AppError{Code: 500, Message: "解压备份文件失败: " + err.Error()}
		}
		defer gzReader.Close()
		// 使用 pipe 将解压数据传给 mysql
		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			io.Copy(pw, gzReader)
		}()

		// 执行还原
		restoreArgs := []string{
			fmt.Sprintf("-u%s", dbUser),
			fmt.Sprintf("-p%s", dbPass),
			fmt.Sprintf("-h%s", dbHost),
			fmt.Sprintf("-P%d", dbPort),
			dbName,
		}
		cmd := exec.Command("mysql", restoreArgs...)
		cmd.Stdin = pr
		if err := cmd.Run(); err != nil {
			return &AppError{Code: 500, Message: "执行还原失败: " + err.Error()}
		}
	} else {
		// 非 gzip 文件直接还原
		restoreArgs := []string{
			fmt.Sprintf("-u%s", dbUser),
			fmt.Sprintf("-p%s", dbPass),
			fmt.Sprintf("-h%s", dbHost),
			fmt.Sprintf("-P%d", dbPort),
			dbName,
		}
		cmd := exec.Command("mysql", restoreArgs...)
		cmd.Stdin = reader
		if err := cmd.Run(); err != nil {
			return &AppError{Code: 500, Message: "执行还原失败: " + err.Error()}
		}
	}

	return nil
}

// ClearData 清除业务数据
func (s *MaintenanceService) ClearData(req *ClearDataRequest) error {
	// 检查是否有10分钟内备份
	hasBackup, _, err := s.CheckRecentBackup()
	if err != nil {
		return &AppError{Code: 500, Message: "检查备份状态失败"}
	}
	if !hasBackup {
		return &AppError{Code: 400, Message: "请先执行数据备份（10分钟内）后再清除数据"}
	}

	// 获取数据表映射
	tableMap := make(map[string][]string)
	for _, info := range s.GetDataTables() {
		tableMap[info.Category] = info.Tables
	}

	// 执行清除
	for _, category := range req.TableCategories {
		tables, ok := tableMap[category]
		if !ok {
			continue
		}
		for _, table := range tables {
			if err := s.db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
				return &AppError{Code: 500, Message: fmt.Sprintf("清除表 %s 失败: %v", table, err)}
			}
		}
	}

	return nil
}
