package excel

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// WarehouseStock 仓库库存信息
type WarehouseStock struct {
	WarehouseName string
	StockQuantity int
}

// ImportRow Excel中导入行数据（一行 = 一个商品，规格颜色自动生成SKU组合）
type ImportRow struct {
	Row             int
	ProductCode     string
	ProductName     string
	Brand           string
	Style           string
	Unit            string
	CategoryName    string  // 品类名称
	ListPrice       float64 // 挂牌价
	CostPrice       float64 // 进货价/成本价
	MinPrice        float64 // 最低售价
	ReferenceCost   float64 // 参考成本
	TotalCostRate   float64 // 成本系数
	WarningStock    int
	Spec            string // 规格（如：三座、四座）
	Color           string // 颜色（如：红色、蓝色）
	Barcode         string
	Series          string // 系列
	SubCategory     string // 类别（A/B/C）
	Warehouse1Name  string // 仓库1名称
	Warehouse1Stock int    // 仓库1库存数量
	Warehouse2Name  string // 仓库2名称
	Warehouse2Stock int    // 仓库2库存数量
	Warehouse3Name  string // 仓库3名称
	Warehouse3Stock int    // 仓库3库存数量
}

// GetWarehouseStocks 获取所有有效仓库库存（仓库名称不为空 AND 库存数量不为0 AND 库存数量>0）
func (row *ImportRow) GetWarehouseStocks() []WarehouseStock {
	var result []WarehouseStock

	// 仓库1：仓库名称不为空 AND 库存数量不为0 AND 库存数量大于0
	if row.Warehouse1Name != "" && row.Warehouse1Stock != 0 && row.Warehouse1Stock > 0 {
		result = append(result, WarehouseStock{
			WarehouseName: row.Warehouse1Name,
			StockQuantity: row.Warehouse1Stock,
		})
	}

	// 仓库2
	if row.Warehouse2Name != "" && row.Warehouse2Stock != 0 && row.Warehouse2Stock > 0 {
		result = append(result, WarehouseStock{
			WarehouseName: row.Warehouse2Name,
			StockQuantity: row.Warehouse2Stock,
		})
	}

	// 仓库3
	if row.Warehouse3Name != "" && row.Warehouse3Stock != 0 && row.Warehouse3Stock > 0 {
		result = append(result, WarehouseStock{
			WarehouseName: row.Warehouse3Name,
			StockQuantity: row.Warehouse3Stock,
		})
	}

	return result
}

// ParseImportFile 解析导入的xlsx文件（单Sheet，每行=一个商品）
func ParseImportFile(file []byte) ([]ImportRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(file))
	if err != nil {
		return nil, fmt.Errorf("无法解析Excel文件: %v", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel文件为空")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取数据失败: %v", err)
	}

	if len(rows) < 2 {
		return nil, nil // 只有表头，无数据
	}

	var result []ImportRow
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || isBlankRow(row) {
			continue
		}

		item := ImportRow{
			Row:             i + 1,
			ProductCode:     getCell(row, 0),
			ProductName:     getCell(row, 1),
			Brand:           getCell(row, 2),
			Style:           getCell(row, 3),
			Unit:            getCell(row, 4),
			CategoryName:    getCell(row, 5),
			ListPrice:       getFloatCell(row, 6),
			CostPrice:       getFloatCell(row, 7),  // 进货价
			MinPrice:        getFloatCell(row, 8),  // 最低售价
			ReferenceCost:   getFloatCell(row, 9),  // 参考成本
			TotalCostRate:   getFloatCell(row, 10), // 成本系数
			WarningStock:    getIntCell(row, 11),
			Spec:            getCell(row, 12),
			Color:           getCell(row, 13),
			Barcode:         getCell(row, 14),
			Series:          getCell(row, 15),
			SubCategory:     getCell(row, 16),
			Warehouse1Name:  getCell(row, 17),    // 仓库1名称
			Warehouse1Stock: getIntCell(row, 18), // 仓库1库存
			Warehouse2Name:  getCell(row, 19),    // 仓库2名称
			Warehouse2Stock: getIntCell(row, 20), // 仓库2库存
			Warehouse3Name:  getCell(row, 21),    // 仓库3名称
			Warehouse3Stock: getIntCell(row, 22), // 仓库3库存
		}
		result = append(result, item)
	}

	return result, nil
}

// GenerateTemplate 生成导入模板xlsx文件（单Sheet）
func GenerateTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// 表头：商品信息(10列) + 规格颜色条码(3列) + 系列类别(2列) + 多仓库库存(6列)
	headers := []string{
		"商品编码", "商品名称", "品牌", "款式", "单位",
		"品类名称", "挂牌价", "进货价", "最低售价", "参考成本", "成本系数", "库存预警",
		"规格", "颜色", "条码", "系列", "类别",
		"仓库1名称", "仓库1库存",
		"仓库2名称", "仓库2库存",
		"仓库3名称", "仓库3库存",
	}
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// 表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E6F7FF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	for col := 1; col <= len(headers); col++ {
		cell, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 示例数据（规格和颜色支持多值，用逗号分隔）
	examples := [][]string{
		{"ZD100001", "真皮沙发A款", "品牌A", "现代简约", "件", "沙发", "5999", "3000", "2500", "3600", "1.2", "10", "三座,四座", "红色,蓝色,米色", "", "现代系列", "A", "主仓库", "50", "副仓库", "30", "备用仓库", "20"},
		{"", "实木餐桌B款", "品牌B", "中式", "张", "餐桌", "3999", "2000", "1800", "2400", "", "5", "1.2米,1.5米", "原木色,胡桃色", "", "古典系列", "B", "主仓库", "40", "", "", "副仓库", "15"},
	}
	for rowIdx, row := range examples {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, val)
		}
	}

	// 列宽
	colWidths := map[string]float64{
		"A": 15, "B": 20, "C": 10, "D": 12, "E": 8,
		"F": 12, "G": 10, "H": 10, "I": 10, "J": 10, "K": 10, "L": 10,
		"M": 15, "N": 20, "O": 18, "P": 15, "Q": 10,
		"R": 12, "S": 10, // 仓库1名称、仓库1库存
		"T": 12, "U": 10, // 仓库2名称、仓库2库存
		"V": 12, "W": 10, // 仓库3名称、仓库3库存
	}
	for col, width := range colWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成模板失败: %v", err)
	}

	return buf.Bytes(), nil
}

// ParseSpecColor 解析规格和颜色，生成SKU组合
// 输入：spec="三座,四座", color="红色,蓝色"
// 输出：[]map[string]string{{"规格":"三座","颜色":"红色"}, {"规格":"三座","颜色":"蓝色"}, ...}
func ParseSpecColor(spec, color string) []map[string]string {
	var result []map[string]string

	specs := splitAndTrim(spec)
	colors := splitAndTrim(color)

	// 笛卡尔积生成组合
	for _, s := range specs {
		for _, c := range colors {
			item := make(map[string]string)
			if s != "" {
				item["规格"] = s
			}
			if c != "" {
				item["颜色"] = c
			}
			if len(item) > 0 {
				result = append(result, item)
			}
		}
	}

	// 如果规格和颜色都为空，返回一个空对象（表示无规格商品）
	if len(result) == 0 {
		result = append(result, map[string]string{})
	}

	return result
}

// AttributesToJSON 将规格颜色map转为JSON字符串
func AttributesToJSON(attrs map[string]string) string {
	if len(attrs) == 0 {
		return "{}"
	}

	var pairs []string
	for k, v := range attrs {
		pairs = append(pairs, fmt.Sprintf("\"%s\":\"%s\"", k, v))
	}
	return "{" + strings.Join(pairs, ",") + "}"
}

// SKUNameFromAttrs 根据属性生成SKU名称
func SKUNameFromAttrs(productName string, attrs map[string]string) string {
	if len(attrs) == 0 {
		return productName
	}

	var parts []string
	// 先加规格，再加颜色
	if spec, ok := attrs["规格"]; ok && spec != "" {
		parts = append(parts, spec)
	}
	if color, ok := attrs["颜色"]; ok && color != "" {
		parts = append(parts, color)
	}

	if len(parts) == 0 {
		return productName
	}
	return productName + "-" + strings.Join(parts, "-")
}

// --- 辅助函数 ---

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func getCell(row []string, index int) string {
	if index < len(row) {
		return strings.TrimSpace(row[index])
	}
	return ""
}

func getFloatCell(row []string, index int) float64 {
	val := getCell(row, index)
	if val == "" {
		return 0
	}
	var f float64
	fmt.Sscanf(val, "%f", &f)
	return f
}

func getIntCell(row []string, index int) int {
	val := getCell(row, index)
	if val == "" {
		return 0
	}
	var i int
	fmt.Sscanf(val, "%d", &i)
	return i
}

func isBlankRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
