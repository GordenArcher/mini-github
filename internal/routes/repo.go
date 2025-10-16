package routes

import (
	"github.com/GordenArcher/mini-github/internal/db"
	"github.com/GordenArcher/mini-github/internal/handlers"
	"github.com/GordenArcher/mini-github/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRepoRoutes(r *gin.RouterGroup, dbConn *db.DB) {
	repoGroup := r.Group("/repos")

	repoGroup.Use(middleware.AuthMiddleware())

	repoGroup.POST("/create", handlers.CreateRepo(dbConn, "/Users/macbookpro/Desktop/mini-github-repos/"))
	repoGroup.GET("/", handlers.ListUserRepos(dbConn))
	repoGroup.GET("/:id", handlers.GetRepo(dbConn))
}
