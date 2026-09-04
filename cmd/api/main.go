package main

import (
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var db *gorm.DB

func main() {
    // بارگذاری متغیرهای محیطی (اگر فایل .env وجود دارد)
    godotenv.Load()

    // اتصال به دیتابیس
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        // برای تست فعلی، این آدرس را به صورت پیش‌فرض قرار دهید (بعداً از لیارا می‌گیرید)
        dsn = "postgresql://root:KrWfSJFMQlPqbGMqMyMwgGad@horizon:5432/postgres?sslmode=disable"
    }

    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("خطا در اتصال به دیتابیس:", err)
    }
    log.Println("✅ اتصال به دیتابیس برقرار شد")

    // راه‌اندازی Gin
    r := gin.Default()

    // فعال‌سازی CORS (برای دسترسی از فرانت‌اند)
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key", "X-Admin-Token"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: false,
        MaxAge:           12 * time.Hour,
    }))

    // مسیرهای اصلی (برای شروع فقط health و branches)
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now()})
    })

    r.GET("/branches", func(c *gin.Context) {
        // اینجا بعداً لیست شعب را از دیتابیس می‌خوانید
        c.JSON(http.StatusOK, gin.H{"message": "لیست شعب به زودی"})
    })

    // اجرای سرور
    port := os.Getenv("SERVER_PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("🚀 سرور Horizon روی پورت %s اجرا شد", port)
    r.Run(":" + port)
}
