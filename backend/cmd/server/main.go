package main

import (
    "ai-hub/backend/internal/api"
    "ai-hub/backend/internal/store"
    "log"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func main() {
    db := store.NewDB("ai_hub.db")
    r := gin.Default()

    // CORS
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge: 12 * time.Hour,
    }))

    h := api.NewHandler(db)
    apiGroup := r.Group("/api")
    {
        apiGroup.GET("/keys", h.ListAPIKeys)
        apiGroup.POST("/keys", h.CreateAPIKey)
        apiGroup.DELETE("/keys/:id", h.DeleteAPIKey)

        apiGroup.GET("/projects", h.ListProjects)
        apiGroup.POST("/projects", h.CreateProject)
        apiGroup.DELETE("/projects/:id", h.DeleteProject)

        // Tasks & Prompts
        apiGroup.GET("/tasks", h.ListTasks)
        apiGroup.POST("/tasks", h.CreateTask)
        apiGroup.DELETE("/tasks/:id", h.DeleteTask)

        // Prompts CRUD for a task
        apiGroup.GET("/tasks/:id/prompts", h.ListPrompts)
        apiGroup.POST("/tasks/:id/prompts", h.CreatePrompt)
        apiGroup.PUT("/tasks/:id/prompts/order", h.ReorderPrompts)
        apiGroup.DELETE("/prompts/:id", h.DeletePrompt)

        // Run a task (execute sequence)
        apiGroup.POST("/tasks/:id/run", h.RunTask)

        // History & Stats
        apiGroup.GET("/history", h.ListHistory)
        apiGroup.GET("/stats", h.Stats)
    }

    log.Println("Starting server on :8080")
    if err := r.Run(":8080"); err != nil {
        log.Fatal(err)
    }
}
