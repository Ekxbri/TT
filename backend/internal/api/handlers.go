package api

import (
    "net/http"
    "strconv"
    "time"
    "strings"

    "ai-hub/backend/internal/models"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type Handler struct {
    DB *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler { return &Handler{DB: db} }

// ---- API Keys
func (h *Handler) ListAPIKeys(c *gin.Context) {
    var keys []models.APIKey
    h.DB.Find(&keys)
    c.JSON(http.StatusOK, keys)
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
    var payload map[string]interface{}
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    k := models.APIKey{}
    if v, ok := payload["provider"]; ok { k.Provider = v.(string) }
    if v, ok := payload["name"]; ok { k.Name = v.(string) }
    if v, ok := payload["value"]; ok { k.Value = v.(string) }
    if v, ok := payload["status"]; ok { k.Status = v.(string) }
    if k.Status == "" { k.Status = "active" }
    h.DB.Create(&k)
    c.JSON(http.StatusCreated, k)
}

func (h *Handler) DeleteAPIKey(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    h.DB.Delete(&models.APIKey{}, id)
    c.Status(http.StatusNoContent)
}

// ---- Projects
func (h *Handler) ListProjects(c *gin.Context) {
    var items []models.Project
    h.DB.Find(&items)
    c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateProject(c *gin.Context) {
    var payload map[string]interface{}
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    p := models.Project{}
    if v, ok := payload["name"]; ok { p.Name = v.(string) }
    if v, ok := payload["api_key"]; ok { p.APIKey = v.(string) }
    if v, ok := payload["status"]; ok { p.Status = v.(string) }
    if p.Status == "" { p.Status = "active" }
    h.DB.Create(&p)
    c.JSON(http.StatusCreated, p)
}

func (h *Handler) DeleteProject(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    h.DB.Delete(&models.Project{}, id)
    c.Status(http.StatusNoContent)
}

// ---- Tasks & Prompts
func (h *Handler) ListTasks(c *gin.Context) {
    var items []models.Task
    h.DB.Preload("Prompts", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" asc") }).Find(&items)
    c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateTask(c *gin.Context) {
    var payload map[string]interface{}
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    t := models.Task{}
    if v, ok := payload["name"]; ok { t.Name = v.(string) }
    if v, ok := payload["description"]; ok { t.Description = v.(string) }
    if v, ok := payload["api_method"]; ok { t.APIMethod = v.(string) }
    if v, ok := payload["version"]; ok { t.Version = v.(string) }
    if v, ok := payload["project_id"]; ok { t.ProjectID = uint(v.(float64)) }
    if t.Status == "" { t.Status = "active" }
    h.DB.Create(&t)
    h.DB.Preload("Prompts", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" asc") }).First(&t, t.ID)
    c.JSON(http.StatusCreated, t)
}

func (h *Handler) DeleteTask(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    h.DB.Delete(&models.Task{}, id)
    c.Status(http.StatusNoContent)
}

func (h *Handler) ListPrompts(c *gin.Context) {
    tid, _ := strconv.Atoi(c.Param("id"))
    var prompts []models.Prompt
    h.DB.Where("task_id = ?", tid).Order("\"order\" asc").Find(&prompts)
    c.JSON(http.StatusOK, prompts)
}

func (h *Handler) CreatePrompt(c *gin.Context) {
    tid, _ := strconv.Atoi(c.Param("id"))
    var payload map[string]interface{}
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    p := models.Prompt{TaskID: uint(tid)}
    if v, ok := payload["name"]; ok { p.Name = v.(string) }
    if v, ok := payload["model"]; ok { p.Model = v.(string) }
    if v, ok := payload["text"]; ok { p.Text = v.(string) }
    var maxOrder int
    h.DB.Model(&models.Prompt{}).Where("task_id = ?", tid).Select("COALESCE(MAX(\"order\"),0)").Scan(&maxOrder)
    p.Order = maxOrder + 1
    h.DB.Create(&p)
    c.JSON(http.StatusCreated, p)
}

func (h *Handler) ReorderPrompts(c *gin.Context) {
    tid, _ := strconv.Atoi(c.Param("id"))
    var payload struct{ Order []uint `json:"order"` }
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    for idx, pid := range payload.Order {
        h.DB.Model(&models.Prompt{}).Where("id = ? AND task_id = ?", pid, tid).Update("order", idx+1)
    }
    c.Status(http.StatusOK)
}

func (h *Handler) DeletePrompt(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    h.DB.Delete(&models.Prompt{}, id)
    c.Status(http.StatusNoContent)
}

// ---- Run Task (simulate LLM provider)
// simple simulation: output = reversed text + timestamp, tokens = word count, cost = tokens * 0.0001
func (h *Handler) RunTask(c *gin.Context) {
    tid, _ := strconv.Atoi(c.Param("id"))
    var task models.Task
    if err := h.DB.Preload("Prompts", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" asc") }).First(&task, tid).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
        return
    }
    projID := task.ProjectID
    var histories []models.PromptHistory
    totalTokens := 0
    totalCost := 0.0
    for _, p := range task.Prompts {
        tokens := countTokensSim(p.Text)
        cost := float64(tokens) * 0.0001
        out := simulateLLMResponse(p.Text)
        ph := models.PromptHistory{PromptID: p.ID, TaskID: task.ID, ProjectID: projID, Input: p.Text, Output: out, Model: p.Model, Tokens: tokens, CostUSD: cost}
        h.DB.Create(&ph)
        histories = append(histories, ph)
        totalTokens += tokens
        totalCost += cost
    }
    c.JSON(http.StatusOK, gin.H{"histories": histories, "total_tokens": totalTokens, "total_cost": totalCost})
}

func countTokensSim(text string) int {
    if strings.TrimSpace(text) == "" { return 0 }
    parts := strings.Fields(text)
    return len(parts)
}

func simulateLLMResponse(text string) string {
    parts := strings.Fields(text)
    for i, j := 0, len(parts)-1; i<j; i,j = i+1, j-1 {
        parts[i], parts[j] = parts[j], parts[i]
    }
    return strings.Join(parts, " ") + " (simulated at " + time.Now().Format(time.RFC3339) + ")"
}

// ---- History & Stats
func (h *Handler) ListHistory(c *gin.Context) {
    var items []models.PromptHistory
    q := h.DB.Order("created_at desc")
    if v := c.Query("project_id"); v != "" {
        q = q.Where("project_id = ?", v)
    }
    if v := c.Query("task_id"); v != "" {
        q = q.Where("task_id = ?", v)
    }
    if v := c.Query("model"); v != "" {
        q = q.Where("model = ?", v)
    }
    q.Find(&items)
    c.JSON(http.StatusOK, items)
}

func (h *Handler) Stats(c *gin.Context) {
    type Row struct {
        ProjectID uint `json:"project_id"`
        TaskID uint `json:"task_id"`
        Model string `json:"model"`
        Tokens int `json:"tokens"`
        Cost float64 `json:"cost"`
    }
    var rows []Row
    h.DB.Table("prompt_histories").Select("project_id, task_id, model, SUM(tokens) as tokens, SUM(cost_usd) as cost").Group("project_id, task_id, model").Scan(&rows)
    c.JSON(http.StatusOK, rows)
}
