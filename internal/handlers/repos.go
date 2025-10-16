package handlers

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/GordenArcher/mini-github/internal/db"
	"github.com/GordenArcher/mini-github/internal/helper/responses"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CreateRepo creates a new repository

func CreateRepo(dbConn *db.DB, basePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		type payload struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
			Visibility  string `json:"visibility"`
		}

		var req payload
		if err := c.ShouldBindJSON(&req); err != nil {
			responses.JSONError(c, 400, "invalid payload")
			return
		}

		userID := c.MustGet("user_id").(uint)

		// Construct repo path
		userFolder := strconv.Itoa(int(userID))
		repoPath := filepath.Join(basePath, userFolder, req.Name+".git")

		// Create directory
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			responses.JSONError(c, 500, "failed to create repo folder")
			return
		}

		// Initialize bare repository
		cmd := exec.Command("git", "init", "--bare")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			responses.JSONError(c, 500, "failed to init bare git repo")
			return
		}

		// Save repo in database
		repo := db.Repository{
			Name:        req.Name,
			Description: req.Description,
			OwnerID:     userID,
			Visibility:  req.Visibility,
			Path:        repoPath,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := dbConn.Create(&repo).Error; err != nil {
			responses.JSONError(c, 500, "failed to save repo")
			return
		}

		responses.JSONSuccess(c, 201, "repository created", gin.H{
			"repo_name": req.Name,
			"clone_url": repoPath,
		})
	}
}

// ListUserRepos lists all repositories for the authenticated user
func ListUserRepos(dbConn *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			responses.JSONError(c, http.StatusUnauthorized, "user not authenticated")
			return
		}
		userID := userIDVal.(uint)

		var repos []db.Repository
		if err := dbConn.Where("owner_id = ?", userID).Find(&repos).Error; err != nil {
			responses.JSONError(c, http.StatusInternalServerError, "cannot fetch repos")
			return
		}

		responses.JSONSuccess(c, http.StatusOK, "ok", repos)
	}
}

// GetRepo fetches repository details including commits and files
func GetRepo(dbConn *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		repoID := c.Param("id")
		var repo db.Repository
		if err := dbConn.Preload("Owner").First(&repo, repoID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "repo not found"})
			return
		}

		userID, ok := c.Get("user_id")
		if repo.Visibility == "private" && (!ok || repo.OwnerID != userID.(uint)) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		r, err := git.PlainOpen(repo.Path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open git repo"})
			return
		}

		ref, err := r.Head()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get HEAD"})
			return
		}

		iter, err := r.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get commit logs"})
			return
		}

		var commits []map[string]interface{}
		_ = iter.ForEach(func(cmt *object.Commit) error {
			commits = append(commits, map[string]interface{}{
				"hash":      cmt.Hash.String(),
				"author":    cmt.Author.Name,
				"email":     cmt.Author.Email,
				"message":   cmt.Message,
				"timestamp": cmt.Author.When,
			})
			return nil
		})

		// Walk repo files
		files := []string{}
		_ = filepath.Walk(repo.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				rel, _ := filepath.Rel(repo.Path, path)
				files = append(files, rel)
			}
			return nil
		})

		c.JSON(http.StatusOK, gin.H{
			"repo":    repo,
			"commits": commits,
			"files":   files,
		})
	}
}
