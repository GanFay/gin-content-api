package router

import (
	"blog/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(h *handlers.Handler) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/ping", h.Ping)
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	r.GET("/auth/refresh", h.Refresh)
	r.POST("/auth/logout", h.Logout)

	authGroup := r.Group("/")
	authGroup.Use(h.AuthMiddleware())
	authGroup.POST("/posts", h.CreateBlog)
	authGroup.PUT("/posts/:id", h.UpdateBlog)
	authGroup.DELETE("/posts/:id", h.DeleteBlog)
	authGroup.GET("/posts", h.GetAllPosts)
	authGroup.GET("/posts/:id", h.GetPoID)

	authGroup.GET("/users/me", h.Me)

	return r
}
