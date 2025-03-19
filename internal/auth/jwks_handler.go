package auth

import (
	"log"
	"sync"

	"github.com/lgm8-measurements-service/internal/httpclient"
)

// JWKSManager handles JWKS key fetch
type JWKSManager struct {
	apiClient *httpclient.HTTPClient
	endpoint  string
	JWKS      *JWKSResponse
	mu        sync.Mutex // Mutex to ensure only one goroutine updates JWKS at a time
}

// NewJWKSManager initializes a new JWKSManager
func NewJWKSManager(client *httpclient.HTTPClient, endpoint string) *JWKSManager {
	return &JWKSManager{
		apiClient: client,
		endpoint:  endpoint,
	}
}

// FetchJWKS retrieves JWKS keys from the auth-service
func (c *JWKSManager) FetchJWKS() error {
	// Lock the mutex to prevent concurrent updates
	c.mu.Lock()
	defer c.mu.Unlock() // Unlock the mutex when the function returns

	var jwks JWKSResponse
	if err := c.apiClient.Get(c.endpoint, &jwks); err != nil {
		// Log the error and return it
		return err
	}

	// Update the JWKS field with the newly fetched data
	c.JWKS = &jwks
	log.Println("JWKS successfully fetched and updated.")
	return nil
}
