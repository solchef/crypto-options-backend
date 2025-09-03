package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var (
	accessTTL  = time.Minute * 180
	refreshTTL = time.Hour * 24 * 7
)

func jwtSecret() []byte {
	// set JWT_SECRET in env for prod
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		sec = "dev-super-secret"
	}
	return []byte(sec)
}

type AccessClaims struct {
	Sub      uint   `json:"sub"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	Sub uint   `json:"sub"`
	JTI string `json:"jti"`
	jwt.RegisteredClaims
}

func GenerateJWT(username string) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret(), nil
	})
}

func NewAccessToken(userID uint, username string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(accessTTL)
	claims := AccessClaims{
		Sub:      userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(jwtSecret())
	return signed, exp, err
}

// returns: token string, jti, expiry, error
func NewRefreshToken(userID uint) (string, string, time.Time, error) {
	now := time.Now()
	exp := now.Add(refreshTTL)
	jti := uuid.NewString()

	claims := RefreshClaims{
		Sub: userID,
		JTI: jti,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        jti,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(jwtSecret())
	return signed, jti, exp, err
}

func ParseRefreshToken(tokenStr string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
