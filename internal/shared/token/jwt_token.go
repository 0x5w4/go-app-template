package token

import (
	"errors"
	"fmt"
	"goapptemp/constant"
	"goapptemp/internal/shared"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type Token interface {
	GenerateAccessToken(userID uint) (string, time.Time, error)
	GenerateRefreshToken(userID uint) (string, time.Time, error)
	VerifyAccessToken(tokenStr string) (*AccessTokenClaims, error)
	VerifyRefreshToken(tokenStr string) (*RefreshTokenClaims, error)
}

type jwtToken struct {
	accessSecretKey      string
	accessTokenDuration  time.Duration
	refreshSecretKey     string
	refreshTokenDuration time.Duration
}

func NewJwtToken(accessSecretKey, refreshSecretKey string, accessTokenDuration, refreshTokenDuration time.Duration) (*jwtToken, error) {
	if len(accessSecretKey) < constant.TokenMinSecretSize || len(refreshSecretKey) < constant.TokenMinSecretSize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", constant.TokenMinSecretSize)
	}

	if accessTokenDuration <= 0 || refreshTokenDuration <= 0 {
		return nil, errors.New("invalid token duration: must be greater than 0")
	}

	return &jwtToken{
		accessSecretKey:      accessSecretKey,
		accessTokenDuration:  accessTokenDuration,
		refreshSecretKey:     refreshSecretKey,
		refreshTokenDuration: refreshTokenDuration,
	}, nil
}

type AccessTokenClaims struct {
	jwt.RegisteredClaims
	UserID uint `json:"user_id"`
}

type RefreshTokenClaims struct {
	jwt.RegisteredClaims
	UserID uint `json:"user_id"`
}

func (j *jwtToken) GenerateAccessToken(userID uint) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.accessTokenDuration)

	uuidStr, err := shared.GenerateUUIDString()
	if err != nil {
		return "", time.Time{}, err
	}

	claims := &AccessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    constant.TokenIssuer,
			Subject:   strconv.FormatUint(uint64(userID), 10),
			ID:        uuidStr,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.accessSecretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (j *jwtToken) GenerateRefreshToken(userID uint) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.refreshTokenDuration)

	uuidStr, err := shared.GenerateUUIDString()
	if err != nil {
		return "", time.Time{}, err
	}

	claims := &RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    constant.TokenIssuer,
			Subject:   strconv.FormatUint(uint64(userID), 10),
			ID:        uuidStr,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.refreshSecretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (j *jwtToken) VerifyAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(j.accessSecretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, errors.New("invalid access token signature")
		}

		return nil, errors.New("invalid access token")
	}

	if !token.Valid {
		return nil, errors.New("invalid access token")
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, errors.New("invalid access token claims")
	}

	return claims, nil
}

func (j *jwtToken) VerifyRefreshToken(tokenStr string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(j.accessSecretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, errors.New("invalid refresh token signature")
		}

		return nil, errors.New("invalid refresh token")
	}

	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return nil, errors.New("invalid refresh token claims")
	}

	return claims, nil
}
