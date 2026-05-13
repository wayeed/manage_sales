package md5

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5Encode 对字符串进行MD5加密
func MD5Encode(password string) string {
	h := md5.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

// MD5Verify 验证密码是否匹配
func MD5Verify(password, encoded string) bool {
	return MD5Encode(password) == encoded
}
