package services

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"SPOTCIN/backend/internal/models"
)

var ErrInvalidCredentials = errors.New("credenciais inválidas")
var ErrEmailTaken = errors.New("email já cadastrado")

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "dev-secret-change-me"
	}
	return []byte(s)
}

func RegisterUser(db *gorm.DB, email, password string) (*models.User, error) {
	var existing models.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:       uuid.NewString(),
		Email:    email,
		Password: string(hash),
	}
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func LoginUser(db *gorm.DB, email, password string) (string, *models.User, error) {
	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return "", nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString(jwtSecret())
	if err != nil {
		return "", nil, err
	}
	return signed, &user, nil
}

func ParseToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return jwtSecret(), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("token inválido")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("claims inválidas")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("sub ausente")
	}
	return sub, nil
}