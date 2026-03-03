package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"zeitpass/internal/models"
	"zeitpass/internal/repository"
)

type AuthService struct {
	userRepo      *repository.UserRepository
	magicLinkRepo *repository.MagicLinkRepository
	emailService  *EmailService
	jwtSecret     []byte
}

func NewAuthService(
	ur *repository.UserRepository,
	mlr *repository.MagicLinkRepository,
	es *EmailService,
) *AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	return &AuthService{
		userRepo:      ur,
		magicLinkRepo: mlr,
		emailService:  es,
		jwtSecret:     []byte(secret),
	}
}

func (s *AuthService) SendMagicLink(email string) error {
	// Find or create user
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// User doesn't exist — auto-create
		uid := generateUserID(8)
		user = &models.User{
			UserID: uid,
			Email:  email,
		}
		if err := s.userRepo.Create(user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Generate token
	rawToken, err := generateSecureToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash for storage
	hash := sha256Hash(rawToken)

	// Store magic link token
	mlToken := &models.MagicLinkToken{
		TokenHash: hash,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.magicLinkRepo.Create(mlToken); err != nil {
		return fmt.Errorf("failed to store magic link: %w", err)
	}

	// Send email
	if err := s.emailService.SendMagicLink(email, rawToken); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

type VerifyResult struct {
	Token string     `json:"token"`
	User  models.User `json:"user"`
}

func (s *AuthService) VerifyMagicLink(rawToken string) (*VerifyResult, error) {
	hash := sha256Hash(rawToken)

	mlToken, err := s.magicLinkRepo.FindValidByHash(hash)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// Mark as used
	if err := s.magicLinkRepo.MarkUsed(mlToken); err != nil {
		return nil, fmt.Errorf("failed to mark token used: %w", err)
	}

	// Look up user
	user, err := s.userRepo.FindByID(mlToken.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Issue JWT
	jwtToken, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &VerifyResult{
		Token: jwtToken,
		User:  *user,
	}, nil
}

type JWTClaims struct {
	UserID   string `json:"userId"`
	UserDbID uint   `json:"userDbId"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

func (s *AuthService) generateJWT(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.UserID,
		UserDbID: user.ID,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "zeitpass",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func sha256Hash(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateUserID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}
