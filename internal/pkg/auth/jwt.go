package auth

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"redditclone/internal/controller/dto"
	"redditclone/internal/domain"
)

type TokenManager struct {
	secret []byte
}

type JwtClaims struct {
	User dto.UserResponse `json:"user"`
	jwt.RegisteredClaims
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: []byte(secret)}
}

func (m *TokenManager) GenerateToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := JwtClaims{
		User: dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Admin:    user.Admin,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		log.Printf("sign token: %v", err)
	}
	return tokenString, err
}

func (m *TokenManager) ParseToken(tokenStr string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*JwtClaims)
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	return &domain.User{
		ID:       claims.User.ID,
		Username: claims.User.Username,
		Admin:    claims.User.Admin,
	}, nil
}
