package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims extends the standard JWT claims with application-specific fields.
type Claims struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	IsSuperAdmin bool  `json:"is_super_admin"`
	jwt.RegisteredClaims
}

// TokenPair holds an access token and its corresponding refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Default durations.
const (
	AccessTokenExpire  = 2 * time.Hour
	RefreshTokenExpire = 7 * 24 * time.Hour
)

// GenerateToken creates an access/refresh token pair for the given user.
// secret is the HMAC signing key; the same key is used for both tokens.
func GenerateToken(userID uint, username string, isSuperAdmin bool, secret []byte) (*TokenPair, error) {
	now := time.Now()

	// Access token
	accessClaims := &Claims{
		UserID:       userID,
		Username:     username,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "access",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessSigned, err := accessToken.SignedString(secret)
	if err != nil {
		return nil, err
	}

	// Refresh token – carries the same user identity but longer TTL.
	refreshClaims := &Claims{
		UserID:       userID,
		Username:     username,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "refresh",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshSigned, err := refreshToken.SignedString(secret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessSigned,
		RefreshToken: refreshSigned,
	}, nil
}

// ParseToken validates and parses a token string, returning the embedded Claims.
func ParseToken(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// Enforce HMAC signing method.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
