package auth

// JWKSResponse represents the structure of the JWKS response
type JWKSResponse struct {
	Keys []map[string]any `json:"keys"`
}
