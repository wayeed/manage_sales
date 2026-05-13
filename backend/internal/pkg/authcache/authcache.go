package authcache

import (
	"fmt"
	"sync"
)

// roleCache 用户角色缓存
type roleCache struct {
	mu    sync.RWMutex
	codes map[int64][]string // userID -> role_codes
}

// permCache 用户权限缓存
type permCache struct {
	mu    sync.RWMutex
	codes map[int64][]string // userID -> permission_codes
}

var Cache = &roleCache{
	codes: make(map[int64][]string),
}

var PermCache = &permCache{
	codes: make(map[int64][]string),
}

// ===== 角色缓存 =====

// Get 获取用户角色编码
func (rc *roleCache) Get(userID int64) ([]string, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	codes, ok := rc.codes[userID]
	return codes, ok
}

// Set 设置用户角色编码
func (rc *roleCache) Set(userID int64, codes []string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.codes[userID] = codes
}

// ===== 权限缓存 =====

// GetPermissions 获取用户权限编码
func (pc *permCache) GetPermissions(userID int64) ([]string, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	codes, ok := pc.codes[userID]
	return codes, ok
}

// SetPermissions 设置用户权限编码
func (pc *permCache) SetPermissions(userID int64, codes []string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.codes[userID] = codes
}

// Invalidate 清除用户所有缓存（角色+权限）
func Invalidate(userID int64) {
	Cache.mu.Lock()
	defer Cache.mu.Unlock()
	delete(Cache.codes, userID)

	PermCache.mu.Lock()
	defer PermCache.mu.Unlock()
	delete(PermCache.codes, userID)
}

// InvalidateRole 清除用户角色缓存
func InvalidateRole(userID int64) {
	Cache.mu.Lock()
	defer Cache.mu.Unlock()
	delete(Cache.codes, userID)
}

// InvalidatePermission 清除用户权限缓存
func InvalidatePermission(userID int64) {
	PermCache.mu.Lock()
	defer PermCache.mu.Unlock()
	delete(PermCache.codes, userID)
}

// InvalidateAll 清除所有缓存
func InvalidateAll() {
	Cache.mu.Lock()
	defer Cache.mu.Unlock()
	Cache.codes = make(map[int64][]string)

	PermCache.mu.Lock()
	defer PermCache.mu.Unlock()
	PermCache.codes = make(map[int64][]string)
}

// permCacheKey 生成权限缓存 key
func permCacheKey(userID int64) string {
	return fmt.Sprintf("perm:%d", userID)
}
