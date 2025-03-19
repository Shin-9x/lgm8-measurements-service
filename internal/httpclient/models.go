package httpclient

// APIError rappresenta un errore standard delle API
type APIError struct {
	Error string `json:"error"`
}
