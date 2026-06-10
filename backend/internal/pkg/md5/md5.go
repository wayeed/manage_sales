package md5

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 对密码进行哈希加密
// cost 参数控制计算复杂度（推荐 bcrypt.DefaultCost = 10）
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword 验证密码是否匹配 bcrypt 哈希值
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// MD5Encode 保留以兼容旧代码，内部升级为 bcrypt
// Deprecated: 请使用 HashPassword 代替。返回值: (hash, error)
func MD5Encode(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	return HashPassword(password)
}

// MD5Verify 保留以兼容旧代码，内部升级为 bcrypt 验证
// Deprecated: 请使用 CheckPassword 代替
func MD5Verify(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}
	return CheckPassword(password, hash)
}
