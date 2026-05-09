package service

import (
	"context"
	"fmt"
	"time"

	"grbac/internal/config"
	"grbac/internal/model"
	"grbac/internal/pkg/crypto"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/jwt"
	"grbac/internal/repository"
	"github.com/redis/go-redis/v9"
)

const tokenBlacklistPrefix = "token:blacklist:"

// LoginResponse holds the tokens returned after a successful login.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// AuthService provides authentication-related operations.
type AuthService struct {
	userRepo    *repository.UserRepo
	redis       *redis.Client
	jwtSecret   []byte
	passwordCfg config.PasswordConfig
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo *repository.UserRepo,
	redis *redis.Client,
	jwtSecret string,
	passwordCfg config.PasswordConfig,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		redis:       redis,
		jwtSecret:   []byte(jwtSecret),
		passwordCfg: passwordCfg,
	}
}

// Login authenticates a user by username and password, returning a token pair on success.
func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, errors.ErrInvalidPassword
	}

	// Check account lock.
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, errors.ErrAccountLocked
	}

	// Check user status.
	if user.Status != 1 {
		return nil, errors.ErrAccountDisabled
	}

	// Verify password.
	if err := crypto.CheckPassword(user.PasswordHash, password); err != nil {
		s.handleLoginFailure(ctx, user)
		return nil, errors.ErrInvalidPassword
	}

	// Login succeeded – reset failure count and clear lock.
	if err := s.resetLoginAttempts(user); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	// Generate tokens.
	tokenPair, err := jwt.GenerateToken(
		uint(user.ID),
		user.Username,
		user.IsSuperAdmin == 1,
		s.jwtSecret,
	)
	if err != nil {
		return nil, errors.ErrInternal.Wrap("生成Token失败")
	}

	return &LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int(jwt.AccessTokenExpire.Seconds()),
	}, nil
}

// RefreshToken validates a refresh token and issues a new token pair.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	claims, err := jwt.ParseToken(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if claims.Subject != "refresh" {
		return nil, errors.ErrTokenInvalid
	}

	user, err := s.userRepo.GetByID(int64(claims.UserID))
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if user.Status != 1 {
		return nil, errors.ErrAccountDisabled
	}

	tokenPair, err := jwt.GenerateToken(
		uint(user.ID),
		user.Username,
		user.IsSuperAdmin == 1,
		s.jwtSecret,
	)
	if err != nil {
		return nil, errors.ErrInternal.Wrap("生成Token失败")
	}

	return &LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int(jwt.AccessTokenExpire.Seconds()),
	}, nil
}

// Logout blacklists the given access token so it can no longer be used.
func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	claims, err := jwt.ParseToken(accessToken, s.jwtSecret)
	if err != nil {
		// Token already invalid – treat as logged out.
		return nil
	}

	// Calculate remaining TTL for the token.
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}

	key := fmt.Sprintf("%s%s", tokenBlacklistPrefix, accessToken)
	return s.redis.Set(ctx, key, "1", ttl).Err()
}

// ChangePassword verifies the old password and sets a new one.
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	if err := crypto.CheckPassword(user.PasswordHash, oldPassword); err != nil {
		return errors.ErrInvalidPassword
	}

	newHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return errors.ErrInternal.Wrap("密码加密失败")
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(user); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}

// IsTokenBlacklisted checks whether the given access token has been blacklisted.
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, accessToken string) bool {
	key := fmt.Sprintf("%s%s", tokenBlacklistPrefix, accessToken)
	val, err := s.redis.Get(ctx, key).Result()
	return err == nil && val == "1"
}

// handleLoginFailure increments the login attempts counter and locks the
// account when the maximum is reached.
func (s *AuthService) handleLoginFailure(_ context.Context, user *model.User) {
	user.LoginAttempts++
	if s.passwordCfg.MaxAttempts > 0 && user.LoginAttempts >= s.passwordCfg.MaxAttempts {
		lockDuration := time.Duration(s.passwordCfg.LockDuration) * time.Minute
		lockUntil := time.Now().Add(lockDuration)
		user.LockedUntil = &lockUntil
	}
	_ = s.userRepo.Update(user)
}

// resetLoginAttempts clears the failure counter and lock on successful login.
func (s *AuthService) resetLoginAttempts(user *model.User) error {
	if user.LoginAttempts == 0 && user.LockedUntil == nil {
		return nil
	}
	user.LoginAttempts = 0
	user.LockedUntil = nil
	return s.userRepo.Update(user)
}
