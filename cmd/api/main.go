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
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ------------------- Database Models -------------------

type Bank struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Wallet    string    `json:"wallet"`
	LicenseID string    `json:"license_id"`
	IsActive  bool      `json:"is_active" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

type Branch struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	BankID    uint      `json:"bank_id"`
	Name      string    `json:"name"`
	City      string    `json:"city"`
	IP        string    `json:"ip"`
	IsOnline  bool      `json:"is_online" gorm:"default:false"`
	Balance   int64     `json:"balance" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
}

type IPWhitelist struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	BankID    uint      `json:"bank_id"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	BranchID  uint      `json:"branch_id"`
	Type      string    `json:"type"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
}

type License struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	ProductID string    `json:"product_id"`
	UserID    string    `json:"user_id"`
	Volume    int       `json:"volume"`
	Root      string    `json:"root"`
	Seed      string    `json:"seed"`
	Signature string    `json:"signature"`
	Status    string    `json:"status"`
	Expiry    time.Time `json:"expiry"`
	CreatedAt time.Time `json:"created_at"`
}

// ------------------- Global Variables -------------------

var db *gorm.DB
var privateKey *ecdsa.PrivateKey
var isSystemActive = false

// ------------------- Helper Functions -------------------

func init() {
	var err error
	privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal("Failed to generate ECDSA key:", err)
	}
	log.Println("✅ ECDSA key generated for license signing")
}

func doSha256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func signData(data []byte, key *ecdsa.PrivateKey) string {
	r, s, err := ecdsa.Sign(rand.Reader, key, data)
	if err != nil {
		log.Println("Failed to sign data:", err)
		return ""
	}
	return hex.EncodeToString(r.Bytes()) + hex.EncodeToString(s.Bytes())
}

func generateLicense(productID, userID string, volume int, durationDays int) License {
	seed := make([]byte, 32)
	_, _ = rand.Read(seed)

	root := doSha256(seed)
	signature := signData([]byte(root), privateKey)

	license := License{
		ID:        "LIC-" + strconv.FormatInt(time.Now().UnixNano(), 10),
		ProductID: productID,
		UserID:    userID,
		Volume:    volume,
		Root:      root,
		Seed:      hex.EncodeToString(seed),
		Signature: signature,
		Status:    "active",
		Expiry:    time.Now().AddDate(0, 0, durationDays),
	}
	return license
}

// ------------------- Main Server -------------------

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://root:KrWfSJFMQlPqbGMqMyMwgGad@horizon:5432/postgres?sslmode=disable"
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("✅ Database connection established")

	if err := db.AutoMigrate(&Bank{}, &Branch{}, &IPWhitelist{}, &Transaction{}, &License{}); err != nil {
		log.Fatal("Failed to auto-migrate database:", err)
	}
	log.Println("✅ Database schema synced")

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key", "X-Admin-Key"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now()})
	})

	r.GET("/api/v1/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": isSystemActive})
	})

	r.POST("/api/v1/admin/init", func(c *gin.Context) {
		if c.GetHeader("X-Admin-Key") != os.Getenv("ADMIN_TOKEN") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin token"})
			return
		}
		isSystemActive = true
		c.JSON(http.StatusOK, gin.H{"status": "System activated"})
	})

	r.POST("/api/v1/admin/shutdown", func(c *gin.Context) {
		if c.GetHeader("X-Admin-Key") != os.Getenv("ADMIN_TOKEN") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin token"})
			return
		}
		isSystemActive = false
		c.JSON(http.StatusOK, gin.H{"status": "System deactivated"})
	})

	r.POST("/api/v1/banks", func(c *gin.Context) {
		if !isSystemActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "System is offline"})
			return
		}
		var bank Bank
		if err := c.ShouldBindJSON(&bank); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&bank).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, bank)
	})

	r.GET("/api/v1/banks", func(c *gin.Context) {
		if !isSystemActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "System is offline"})
			return
		}
		var banks []Bank
		if err := db.Find(&banks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, banks)
	})

	r.POST("/api/v1/branches", func(c *gin.Context) {
		if !isSystemActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "System is offline"})
			return
		}
		var branch Branch
		if err := c.ShouldBindJSON(&branch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&branch).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, branch)
	})

	r.GET("/api/v1/branches", func(c *gin.Context) {
		if !isSystemActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "System is offline"})
			return
		}
		var branches []Branch
		if err := db.Find(&branches).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, branches)
	})

	r.POST("/api/v1/ip/whitelist", func(c *gin.Context) {
		if !isSystemActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "System is offline"})
			return
		}
		var ip IPWhitelist
		if err := c.ShouldBindJSON(&ip); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&ip).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, ip)
	})

	r.POST("/api/v1/transactions", func(c *gin.Context) {
		if !isSystemActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "System is offline"})
			return
		}
		var tx Transaction
		if err := c.ShouldBindJSON(&tx); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		tx.Hash = doSha256([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
		if err := db.Create(&tx).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, tx)
	})

	r.GET("/api/v1/transactions", func(c *gin.Context) {
		if !isSystemActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "System is offline"})
			return
		}
		var txs []Transaction
		if err := db.Find(&txs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, txs)
	})

	r.POST("/api/v1/license/generate", func(c *gin.Context) {
		if !isSystemActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "System is offline"})
			return
		}
		var req struct {
			ProductID string `json:"product_id"`
			UserID    string `json:"user_id"`
			Volume    int    `json:"volume"`
			Duration  int    `json:"duration"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		license := generateLicense(req.ProductID, req.UserID, req.Volume, req.Duration)
		if err := db.Create(&license).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, license)
	})

	r.POST("/api/v1/license/verify", func(c *gin.Context) {
		var req struct {
			LicenseID string `json:"license_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var license License
		if err := db.Where("id = ?", req.LicenseID).First(&license).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"valid": false, "error": "License not found"})
			return
		}
		valid := license.Status == "active" && time.Now().Before(license.Expiry)
		c.JSON(http.StatusOK, gin.H{"valid": valid, "expiry": license.Expiry, "user_id": license.UserID, "product_id": license.ProductID})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Horizon Core Engine running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
