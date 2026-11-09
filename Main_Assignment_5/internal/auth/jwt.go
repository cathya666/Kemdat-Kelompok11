package auth

import (
	"aaa_bekend/internal/lmodel"
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var privateKey *rsa.PrivateKey
var publicKey *rsa.PublicKey

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func LoadKeys() error {

	privKeyData, err := os.ReadFile("keys/private.pem")
	if err != nil {
		return err
	}

	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privKeyData)
	if err != nil {
		return err
	}

	pubKeyData, err := os.ReadFile("keys/public.pem")
	if err != nil {
		return err
	}

	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(pubKeyData)
	if err != nil {
		return err
	}

	return nil
}

func GenerateToken(user *lmodel.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(120 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}
func keyFunc(token *jwt.Token) (interface{}, error) {
	_, ok := token.Method.(*jwt.SigningMethodRSA)
	if !ok {
		return nil, errors.New("Unexpected signing method")
	}
	return publicKey, nil
}
func ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, keyFunc)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("Invalid Token")
}
