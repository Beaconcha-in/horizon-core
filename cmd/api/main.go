package main

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type Bank struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name"`
    Wallet    string    `json:"wallet"`
    LicenseID string    `json:"license_id"`
    IsActive  bool      `json:"is_active" gorm:"default:false"`
    CreatedAt time.Time `json:"created_at"`
}

type IPWhitelist struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    BankID    uint      `json:"bank_id"`
    IP        string    `json:"ip"`
    CreatedAt time.Time `json:"created_at"`
}

var db *gorm.DB
var isSystemActive = false
var privateKey *ecdsa.PrivateKey

func init() {
    privateKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func main() {
    os.Setenv("ADMIN_TOKEN", "test-admin-token")
    os.Setenv("API_KEY", "test-api-key")

    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgresql://root:KrWfSJFMQlPqbGMqMyMwgGad@horizon:5432/postgres?sslmode=disable"
    }

    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("DB connection error:", err)
    }
    db.AutoMigrate(&Bank{}, &IPWhitelist{})
    log.Println("DB connected")

    r := gin.Default()
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key", "X-Admin-Key"},
        AllowCredentials: false,
        MaxAge:           12 * time.Hour,
    }))

    r.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "Horizon Backend"})
    })

    r.GET("/api/v1/status", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": isSystemActive, "service": "Horizon Backend"})
    })

    r.POST("/api/v1/admin/init", func(c *gin.Context) {
        if c.GetHeader("X-Admin-Key") != os.Getenv("ADMIN_TOKEN") {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }
        isSystemActive = true
        c.JSON(http.StatusOK, gin.H{"status": "active"})
    })

    r.POST("/api/v1/admin/shutdown", func(c *gin.Context) {
        if c.GetHeader("X-Admin-Key") != os.Getenv("ADMIN_TOKEN") {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }
        isSystemActive = false
        c.JSON(http.StatusOK, gin.H{"status": "inactive"})
    })

    r.POST("/api/v1/banks", func(c *gin.Context) {
        if !isSystemActive {
            c.JSON(http.StatusForbidden, gin.H{"error": "System is inactive"})
            return
        }
        var bank Bank
        if err := c.ShouldBindJSON(&bank); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        db.Create(&bank)
        c.JSON(http.StatusCreated, bank)
    })

    r.GET("/api/v1/banks", func(c *gin.Context) {
        if !isSystemActive {
            c.JSON(http.StatusForbidden, gin.H{"error": "System is inactive"})
            return
        }
        var banks []Bank
        db.Find(&banks)
        c.JSON(http.StatusOK, banks)
    })

    r.POST("/api/v1/ip/whitelist", func(c *gin.Context) {
        if !isSystemActive {
            c.JSON(http.StatusForbidden, gin.H{"error": "System is inactive"})
            return
        }
        var ip IPWhitelist
        if err := c.ShouldBindJSON(&ip); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        db.Create(&ip)
        c.JSON(http.StatusCreated, ip)
    })

    r.POST("/api/v1/license/verify", func(c *gin.Context) {
        var req struct {
            LicenseID string `json:"license_id"`
        }
        c.ShouldBindJSON(&req)
        hash := sha256.Sum256([]byte(req.LicenseID))
        signature := "SIG:" + hex.EncodeToString(hash[:])
        c.JSON(http.StatusOK, gin.H{"valid": true, "signature": signature})
    })

    port := os.Getenv("SERVER_PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Server running on port %s", port)
    r.Run(":" + port)
}