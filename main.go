package main

import (
	"log"
	"os"
	"path/filepath"

	"inkread/api"
	"inkread/services"
	"inkread/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./inkread.db"
	}

	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		log.Fatalf("无效的上传目录: %v", err)
	}
	if err := os.MkdirAll(absUploadDir, 0755); err != nil {
		log.Fatalf("无法创建上传目录: %v", err)
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}
	defer store.Close()

	bookService := services.NewBookService(store, absUploadDir)

	openAIKey := os.Getenv("OPENAI_API_KEY")
	openAIModel := os.Getenv("OPENAI_MODEL")
	if openAIModel == "" {
		openAIModel = "gpt-3.5-turbo"
	}
	aiService := services.NewAIService(openAIKey, openAIModel)

	handlers := api.NewHandlers(bookService, aiService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Static("/uploads", uploadDir)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	handlers.RegisterRoutes(r)

	log.Printf("InkRead 服务启动于 http://localhost:%s", port)
	log.Printf("上传目录: %s", absUploadDir)
	log.Printf("数据库: %s", dbPath)
	if openAIKey != "" {
		log.Printf("AI 总结: 已启用 (模型: %s)", openAIModel)
	} else {
		log.Printf("AI 总结: 模拟模式")
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
