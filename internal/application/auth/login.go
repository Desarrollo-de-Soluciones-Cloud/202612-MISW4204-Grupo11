package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials is returned when email/password do not match a stored user.
var ErrInvalidCredentials = errors.New("invalid credentials")

type LoginService struct {
	Users  ports.UserRepository
	Secret []byte
}

type LoginResult struct {
	Token string
	User  domain.User
}

func (service *LoginService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	var result LoginResult
	if len(service.Secret) == 0 {
		return result, errors.New("JWT signing secret is not configured")
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	credentials, findErr := service.Users.FindCredentialsByEmail(ctx, normalizedEmail)
	if findErr != nil {
		return result, findErr
	}
	if credentials == nil {
		return result, ErrInvalidCredentials
	}
	if compareErr := bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(password)); compareErr != nil {
		return result, ErrInvalidCredentials
	}

	signedToken, signErr := signToken(credentials.ID, credentials.Roles, service.Secret, 24*time.Hour)
	if signErr != nil {
		return result, signErr
	}

	result.Token = signedToken
	result.User = domain.User{
		ID:    credentials.ID,
		Name:  credentials.Name,
		Email: credentials.Email,
		Roles: credentials.Roles,
	}
	return result, nil
}

type tokenClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func signToken(userID int64, roles []string, secret []byte, ttl time.Duration) (string, error) {
	claims := tokenClaims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken validates the JWT and returns the user id and role names from claims.
func ParseToken(tokenString string, secret []byte) (userID int64, roles []string, err error) {
	var parsed tokenClaims
	_, err = jwt.ParseWithClaims(tokenString, &parsed, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return 0, nil, err
	}
	subject, subjectErr := parsed.GetSubject()
	if subjectErr != nil {
		return 0, nil, subjectErr
	}
	userID, err = strconv.ParseInt(subject, 10, 64)
	if err != nil {
		return 0, nil, err
	}
	return userID, parsed.Roles, nil
}
