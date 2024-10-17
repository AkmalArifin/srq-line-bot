package utils

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwk"
)

var kid string
var channelID string
var privateKeyFile string

func GenerateJWTToken() (string, error) {
	kid = os.Getenv("LINE_KID")
	channelID = os.Getenv("LINE_CHANNEL_ID")
	privateKeyFile = os.Getenv("PRIVATE_KEY_FILE")

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":       channelID,
		"sub":       channelID,
		"aud":       "https://api.line.me/",
		"exp":       time.Now().Add(30 * time.Minute).Unix(),
		"token_exp": 86400,
	})
	token.Header["kid"] = kid

	privateKey, err := getPrivateKey(privateKeyFile)

	if err != nil {
		return "", err
	}

	tokenString, err := token.SignedString(privateKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func getPrivateKey(filename string) (*rsa.PrivateKey, error) {
	// get private key
	dir, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	filePath := dir + "/" + filename
	data, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	key, err := jwk.ParseKey(data)

	if err != nil {
		return nil, err
	}

	var rawkey interface{}
	err = key.Raw(&rawkey)

	if err != nil {
		return nil, err
	}

	rsa, ok := rawkey.(*rsa.PrivateKey)

	if !ok {
		return nil, errors.New("expected rsa key")
	}

	return rsa, nil
}
