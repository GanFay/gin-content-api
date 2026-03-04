package router

import (
	"blog/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(h *handlers.Handler) *gin.Engine {
	r := gin.Default()
	r.GET("/ping", h.Ping)
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	authGroup := r.Group("/")
	authGroup.Use(h.AuthMiddleware())
	authGroup.POST("/posts", h.CreateBlog)
	authGroup.PUT("/posts/:id", h.UpdateBlog)
	authGroup.DELETE("/posts/:id", h.DeleteBlog)
	authGroup.GET("/posts", h.GetAllPosts)
	authGroup.GET("/posts/:id", h.GetPoID)

	return r
}
