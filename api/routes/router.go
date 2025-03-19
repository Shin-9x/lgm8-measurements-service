package routes

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lgm8-measurements-service/api/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	trustedProxies := getTrustedProxiesFromEnv()
	if trustedProxies != nil {
		r.SetTrustedProxies(trustedProxies)
	}

	r.Use(middleware.Logger())

	return r
}

func getTrustedProxiesFromEnv() []string {
	trustedProxiesStr := os.Getenv("GIN_TRUSTED_PROXIES")

	if len(trustedProxiesStr) > 0 && trustedProxiesStr != "" {
		proxies := strings.Split(trustedProxiesStr, ",")

		var res []string
		res = append(res, proxies...)
		return res
	}

	return nil
}
