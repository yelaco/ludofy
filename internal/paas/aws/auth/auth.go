package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type jwk struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

// Load Cognito public keys
func LoadCognitoPublicKeys(url string) (map[string]*rsa.PublicKey, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching Cognito public keys:", err)
		return nil, fmt.Errorf("failed to fetch public keys: %w ", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var jwks jwks
	if err = json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal jwks: %w", err)
	}

	cognitoPublicKeys := make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		// Decode Base64URL (without padding) `n` and `e`
		nBytes, _ := decodeBase64URL(key.N)
		eBytes, _ := decodeBase64URL(key.E)

		// Convert to big.Int and integer
		n := new(big.Int).SetBytes(nBytes)
		e := int(new(big.Int).SetBytes(eBytes).Int64())

		// Construct RSA Public Key
		cognitoPublicKeys[key.Kid] = &rsa.PublicKey{N: n, E: e}
	}

	return cognitoPublicKeys, nil
}

// Decode Base64URL without padding
func decodeBase64URL(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// Validate JWT
func ValidateJwt(
	tokenString string,
	cognitoPublicKeys map[string]*rsa.PublicKey,
) (
	*jwt.Token,
	error,
) {
	if cognitoPublicKeys == nil {
		return nil, fmt.Errorf("cognito public keys not loaded")
	}
	issuer := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s",
		os.Getenv("AWS_REGION"),
		os.Getenv("COGNITO_USER_POOL_ID"),
	)
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New("invalid token: missing kid")
			}
			if key, found := cognitoPublicKeys[kid]; found {
				return key, nil
			}
			return nil, errors.New("invalid token: unknown kid")
		},
		jwt.WithIssuer(issuer),
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func MustAuth(authorizer map[string]interface{}) string {
	jwt, ok := authorizer["jwt"].(map[string]interface{})
	if !ok {
		panic("no jwt")
	}
	v, exists := jwt["claims"]
	if !exists {
		panic("no authorizer claims")
	}
	claims, ok := v.(map[string]interface{})
	if !ok {
		panic("claims must be of type map")
	}
	userId, ok := claims["sub"].(string)
	if !ok {
		panic("invalid sub")
	}
	return userId
}
