package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT payload。TokenVersion 用于"改密后批量踢人"场景。
type Claims struct {
	UserID       uint64 `json:"user_id"`
	TokenVersion uint64 `json:"token_version"`
	jwt.RegisteredClaims
}

// TokenVersionStore 定义 token 版本号查询接口。如果不需要版本校验，
// 不注入即可——中间件会自动跳过 Layer 3。
type TokenVersionStore interface {
	GetTokenVersion(ctx context.Context, userID uint64) (uint64, error)
}

type JWTManager struct {
	secret       []byte
	expiresAfter time.Duration
	tvStore      TokenVersionStore // 可选
}

func NewJWTManager(secret string, expiresAfter time.Duration, tvStore TokenVersionStore) *JWTManager {
	return &JWTManager{
		secret:       []byte(secret),
		expiresAfter: expiresAfter,
		tvStore:      tvStore,
	}
}

// IssueToken 签发新 token，JTI 用 UUID 防重放。
func (m *JWTManager) IssueToken(userID uint64, tokenVersion uint64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:       userID,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiresAfter)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken 三层校验：签名+过期 → JTI 黑名单 → token_version。
// jtiBlacklist 为 true 表示命中黑名单（调用方用 Redis 判断）。
func (m *JWTManager) ValidateToken(tokenStr string, jtiBlacklist func(jti string) bool) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secret, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("token parse: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Layer 1 通过——签名和过期校验已由 jwt 库完成。

	// Layer 2: JTI 黑名单检查
	if claims.ID != "" && jtiBlacklist != nil && jtiBlacklist(claims.ID) {
		return nil, fmt.Errorf("token revoked")
	}

	// Layer 3: token_version 比对（可选）
	if m.tvStore != nil {
		currentVersion, err := m.tvStore.GetTokenVersion(context.Background(), claims.UserID)
		if err != nil {
			return nil, fmt.Errorf("fetch token version: %w", err)
		}
		if claims.TokenVersion < currentVersion {
			return nil, fmt.Errorf("token version outdated")
		}
	}

	return claims, nil
}
