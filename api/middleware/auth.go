package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lgm8-measurements-service/internal/auth"
)

// Authenticate verifies and validates the JWT token with Keycloak
func Authenticate(jf *auth.JWKSManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respondUnauthorized(c, "Missing Authorization header")
			return
		}

		// Extract the token from the "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			respondUnauthorized(c, "Invalid Authorization header format")
			return
		}
		tokenString := tokenParts[1]

		// Validate the token
		_, err := validateToken(tokenString, jf)
		if err != nil {
			respondUnauthorized(c, err.Error())
			return
		}

		log.Printf("Token validated successfully")

		// Continue to the next handler
		c.Next()
	}
}

// validateToken handles parsing and validating the token, including claims
func validateToken(tokenString string, jf *auth.JWKSManager) (*jwt.MapClaims, error) {
	token, err := parseAndValidateToken(tokenString, jf)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	if err := verifyClaims(claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

// verifyClaims checks standard claims like exp, nbf, and iat
func verifyClaims(claims jwt.MapClaims) error {
	now := time.Now()

	if exp, ok := claims["exp"].(float64); ok {
		if now.After(time.Unix(int64(exp), 0)) {
			return fmt.Errorf("token is expired")
		}
	} else {
		return fmt.Errorf("missing or invalid exp claim")
	}

	if nbf, ok := claims["nbf"].(float64); ok {
		if now.Before(time.Unix(int64(nbf), 0)) {
			return fmt.Errorf("token is not yet valid")
		}
	}

	if iat, ok := claims["iat"].(float64); ok {
		if now.Before(time.Unix(int64(iat), 0)) {
			return fmt.Errorf("token issued in the future, possible clock skew issue")
		}
	} else {
		return fmt.Errorf("missing or invalid iat claim")
	}

	return nil
}

// parseAndValidateToken parses and validates the JWT token using JWKS
func parseAndValidateToken(tokenString string, jf *auth.JWKSManager) (*jwt.Token, error) {
	const maxRetries = 3

	for retry := 0; retry <= maxRetries; retry++ {
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return verifyKey(token, *jf.JWKS)
		})

		if err == nil && token.Valid {
			return token, nil // Valid token, exit the loop
		}

		// Check if the error is due to a key not found or an invalid signature
		if errors.Is(err, jwt.ErrSignatureInvalid) || errors.Is(err, jwt.ErrTokenUnverifiable) {
			if err := jf.FetchJWKS(); err != nil {
				return nil, err
			}
			time.Sleep(time.Duration(retry) * time.Second) // Exponential backoff
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf("failed to validate token after [%d] retries", maxRetries)
}

// verifyKey verifies the JWT token key using JWKS
func verifyKey(token *jwt.Token, jwks auth.JWKSResponse) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing kid in token header")
	}

	for _, key := range jwks.Keys {
		if key["kid"] == kid {
			x5c, ok := key["x5c"].([]any)
			if !ok || len(x5c) == 0 {
				return nil, fmt.Errorf("invalid x5c format in JWKS")
			}

			certBytes, err := base64.StdEncoding.DecodeString(x5c[0].(string))
			if err != nil {
				return nil, fmt.Errorf("failed to decode base64 certificate: %w", err)
			}

			cert, err := x509.ParseCertificate(certBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}

			rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
			if !ok {
				return nil, fmt.Errorf("public key is not RSA")
			}

			return rsaPublicKey, nil
		}
	}

	return nil, fmt.Errorf("key not found in JWKS")
}

// respondUnauthorized sends an HTTP 401 response with the specified error message
func respondUnauthorized(c *gin.Context, message string) {
	log.Println(message)
	c.JSON(http.StatusUnauthorized, gin.H{"error": message})
	c.Abort()
}
