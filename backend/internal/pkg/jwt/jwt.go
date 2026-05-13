package jwt

import (
	"errors"
	"time"

	"furniture-commission/configs"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义JWT声明
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID int64) (string, error) {
	cfg := configs.GlobalConfig
	if cfg == nil {
		return "", errors.New("config not initialized")
	}

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.ExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*Claims, error) {
	cfg := configs.GlobalConfig
	if cfg == nil {
		return nil, errors.New("config not initialized")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
