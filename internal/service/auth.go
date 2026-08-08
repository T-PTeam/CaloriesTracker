package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/root1/calories-tracker/internal/domain"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAlreadyRegistered  = errors.New("account already registered")
	ErrInvalidLinkCode    = errors.New("invalid or expired link code")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidEmail       = errors.New("invalid email")
)

type AuthStore interface {
	EnsureUser(ctx context.Context, telegramID int64) (domain.User, error)
	FindUserByID(ctx context.Context, userID int64) (domain.User, error)
	FindUserByEmail(ctx context.Context, email string) (domain.User, error)
	FindUserByTelegramID(ctx context.Context, telegramID int64) (domain.User, error)
	SetUserCredentials(ctx context.Context, userID int64, email, passwordHash string) (domain.User, error)
	UpdateUserLanguage(ctx context.Context, userID int64, language string) (domain.User, error)
	CreateLinkCode(ctx context.Context, userID int64, code string, expiresAt time.Time) (domain.LinkCode, error)
	FindLinkCodeUser(ctx context.Context, code string) (domain.User, error)
	DeleteLinkCode(ctx context.Context, code string) error
}

type AuthService struct {
	store     AuthStore
	jwtSecret []byte
	tokenTTL  time.Duration
}

type TokenClaims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type AuthResult struct {
	Token string     `json:"token"`
	User  PublicUser `json:"user"`
}

type PublicUser struct {
	ID         int64  `json:"id"`
	Email      string `json:"email"`
	TelegramID int64  `json:"telegram_id"`
	Language   string `json:"language"`
}

func NewAuthService(store AuthStore, jwtSecret string, tokenTTL time.Duration) *AuthService {
	if tokenTTL <= 0 {
		tokenTTL = 30 * 24 * time.Hour
	}
	return &AuthService{
		store:     store,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  tokenTTL,
	}
}

func (s *AuthService) CreateWebLinkCode(ctx context.Context, telegramID int64) (string, error) {
	user, err := s.store.EnsureUser(ctx, telegramID)
	if err != nil {
		return "", err
	}
	code, err := randomDigits(6)
	if err != nil {
		return "", err
	}
	_, err = s.store.CreateLinkCode(ctx, user.ID, code, time.Now().UTC().Add(15*time.Minute))
	if err != nil {
		return "", err
	}
	return code, nil
}

func (s *AuthService) Register(ctx context.Context, email, password, linkCode string) (AuthResult, error) {
	email = normalizeEmail(email)
	if !looksLikeEmail(email) {
		return AuthResult{}, ErrInvalidEmail
	}
	if len(password) < 8 {
		return AuthResult{}, ErrWeakPassword
	}
	linkCode = strings.TrimSpace(linkCode)
	if linkCode == "" {
		return AuthResult{}, ErrInvalidLinkCode
	}

	user, err := s.store.FindLinkCodeUser(ctx, linkCode)
	if err != nil {
		return AuthResult{}, ErrInvalidLinkCode
	}
	if user.HasCredentials() {
		return AuthResult{}, ErrAlreadyRegistered
	}

	if _, err := s.store.FindUserByEmail(ctx, email); err == nil {
		return AuthResult{}, ErrAlreadyRegistered
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	user, err = s.store.SetUserCredentials(ctx, user.ID, email, string(hash))
	if err != nil {
		return AuthResult{}, err
	}
	_ = s.store.DeleteLinkCode(ctx, linkCode)

	token, err := s.issueToken(user)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Token: token, User: toPublicUser(user)}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (AuthResult, error) {
	email = normalizeEmail(email)
	user, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	if !user.HasCredentials() {
		return AuthResult{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	token, err := s.issueToken(user)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Token: token, User: toPublicUser(user)}, nil
}

func (s *AuthService) ParseToken(tokenString string) (TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return TokenClaims{}, err
	}
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return TokenClaims{}, fmt.Errorf("invalid token")
	}
	return *claims, nil
}

func (s *AuthService) GetPublicUser(ctx context.Context, userID int64) (PublicUser, error) {
	user, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		return PublicUser{}, err
	}
	return toPublicUser(user), nil
}

func (s *AuthService) LanguageForTelegram(ctx context.Context, telegramID int64) string {
	user, err := s.store.FindUserByTelegramID(ctx, telegramID)
	if err != nil {
		return domain.LangUK
	}
	if user.Language == "" {
		return domain.LangUK
	}
	return domain.NormalizeLanguage(user.Language)
}

func (s *AuthService) SetLanguage(ctx context.Context, telegramID int64, language string) (domain.User, error) {
	if !domain.IsValidLanguage(language) {
		return domain.User{}, fmt.Errorf("invalid language")
	}
	user, err := s.store.EnsureUser(ctx, telegramID)
	if err != nil {
		return domain.User{}, err
	}
	return s.store.UpdateUserLanguage(ctx, user.ID, language)
}

func (s *AuthService) SetLanguageByUserID(ctx context.Context, userID int64, language string) (PublicUser, error) {
	if !domain.IsValidLanguage(language) {
		return PublicUser{}, fmt.Errorf("invalid language")
	}
	user, err := s.store.UpdateUserLanguage(ctx, userID, language)
	if err != nil {
		return PublicUser{}, err
	}
	return toPublicUser(user), nil
}

func (s *AuthService) issueToken(user domain.User) (string, error) {
	now := time.Now().UTC()
	claims := TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func toPublicUser(user domain.User) PublicUser {
	lang := user.Language
	if lang == "" {
		lang = domain.LangUK
	}
	return PublicUser{
		ID:         user.ID,
		Email:      user.Email,
		TelegramID: user.TelegramID,
		Language:   domain.NormalizeLanguage(lang),
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func looksLikeEmail(email string) bool {
	at := strings.IndexByte(email, '@')
	if at <= 0 || at == len(email)-1 {
		return false
	}
	return strings.Contains(email[at+1:], ".")
}

func randomDigits(n int) (string, error) {
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate code: %w", err)
		}
		b.WriteByte(byte('0' + v.Int64()))
	}
	return b.String(), nil
}
