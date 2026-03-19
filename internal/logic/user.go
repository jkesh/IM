package logic

import (
	"IM/internal/model"
	"IM/internal/storage/db"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidUserInput     = errors.New("username must be 3-32 chars and password must be at least 6 chars")
	ErrUserAlreadyExists    = errors.New("username already exists")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrInvalidToken         = errors.New("invalid token")

	jwtKey   = []byte("change-this-secret")
	tokenTTL = 24 * time.Hour
)

type MyClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func SetJWTSecret(secret string) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return
	}
	jwtKey = []byte(secret)
}

func SetTokenTTL(ttl time.Duration) {
	if ttl > 0 {
		tokenTTL = ttl
	}
}

func Register(username, password string) (*model.User, error) {
	username, password, err := normalizeCredentials(username, password)
	if err != nil {
		return nil, err
	}
	if err := ensureDBReady(); err != nil {
		return nil, err
	}

	var existing model.User
	err = db.DB.Where("username = ?", username).First(&existing).Error
	switch {
	case err == nil:
		return nil, ErrUserAlreadyExists
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username: username,
		Password: string(hashedPassword),
	}
	if err := db.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func Login(username, password string) (string, *model.User, error) {
	username, password, err := normalizeCredentials(username, password)
	if err != nil {
		return "", nil, err
	}
	if err := ensureDBReady(); err != nil {
		return "", nil, err
	}

	var user model.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrAuthenticationFailed
		}
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrAuthenticationFailed
	}

	token, err := GenerateToken(strconv.FormatUint(uint64(user.ID), 10))
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

func GetUserByID(userID string) (*model.User, error) {
	if err := ensureDBReady(); err != nil {
		return nil, err
	}

	id, err := strconv.ParseUint(strings.TrimSpace(userID), 10, 64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuthenticationFailed
		}
		return nil, err
	}
	return &user, nil
}

func GenerateToken(userID string) (string, error) {
	now := time.Now()
	claims := &MyClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateToken(tokenString string) (string, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return "", ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return jwtKey, nil
	})
	if err != nil {
		return "", ErrInvalidToken
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", ErrInvalidToken
}

func normalizeCredentials(username, password string) (string, string, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	if len(username) < 3 || len(username) > 32 || len(password) < 6 || len(password) > 72 {
		return "", "", ErrInvalidUserInput
	}
	return username, password, nil
}

func ensureDBReady() error {
	if db.DB == nil {
		return errors.New("database is not initialized")
	}
	return nil
}
