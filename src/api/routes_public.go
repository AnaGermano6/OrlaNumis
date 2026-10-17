package main

import (
	"github.com/gin-gonic/gin"
)

// registerPublicRoutes registers the unauthenticated /api endpoints: login,
// registration and token refresh.
func registerPublicRoutes(api *gin.RouterGroup, d *appDeps) {
	api.POST("/auth/register", d.authRateLimit, d.authHandler.Register)
	api.POST("/auth/login", d.authRateLimit, d.authHandler.Login)
	api.POST("/auth/refresh", d.authRateLimit, d.authHandler.Refresh)
}
