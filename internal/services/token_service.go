package services

import (
	"fmt"
	"time"

	"github.com/fuzzy-toozy/gophermart/internal/models"
	"gopkg.in/go-jose/go-jose.v2"
	"gopkg.in/go-jose/go-jose.v2/jwt"
)

type AppClaims struct {
	jwt.Claims
}

type TokenService interface {
	Generate(user *models.User) (token string, err error)
	Validate(token string) (claims jwt.Claims, err error)
	Duration() time.Duration
}

type DefaultTokenService struct {
	signingKey    jose.SigningKey
	tokenLifetime time.Duration
}

func (s *DefaultTokenService) Duration() time.Duration {
	return s.tokenLifetime
}

func (s *DefaultTokenService) Generate(user *models.User) (token string, err error) {
	signer, err := jose.NewSigner(s.signingKey, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return "", err
	}

	builder := jwt.Signed(signer)
	currentTime := time.Now()
	token, err = builder.Claims(jwt.Claims{
		Issuer:    "auth-service",
		Subject:   user.Username,
		Expiry:    jwt.NewNumericDate(currentTime.Add(s.tokenLifetime)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}).CompactSerialize()

	if err != nil {
		return "", err
	}
	return token, nil
}

func (service *DefaultTokenService) Validate(token string) (claims jwt.Claims, err error) {
	appClaims := AppClaims{}
	jwt, err := jwt.ParseSigned(token)
	if err != nil {
		return appClaims.Claims, fmt.Errorf("failed to parse token. Token is invalid: %w", err)
	}

	err = jwt.Claims(service.signingKey.Key, &appClaims)

	if err != nil {
		return appClaims.Claims, fmt.Errorf("failed to verify token signature: %w", err)
	}

	return appClaims.Claims, nil
}

func NewTokenService(signingKeyBytes []byte, tokenLifetime time.Duration) TokenService {
	signingKey := jose.SigningKey{
		Algorithm: jose.HS256,
		Key:       signingKeyBytes,
	}
	return &DefaultTokenService{
		signingKey:    signingKey,
		tokenLifetime: tokenLifetime,
	}
}
