package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Validator struct {
	Index            int     `json:"index"`
	PublicKey        string  `json:"public_key"`
	CurrentBalance   float64 `json:"current_balance"`
	EffectiveBalance float64 `json:"effective_balance"`
	ActivationDate   string  `json:"activation_date"`
	Income1d         float64 `json:"income_1d"`
	Income7d         float64 `json:"income_7d"`
	Income14d        float64 `json:"income_14d"`
	Income30d        float64 `json:"income_30d"`
	TotalIncome      float64 `json:"total_income"`
}

// GetValidators returns the list of validators from the JSON file
func GetValidators() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := os.Open("internal/data/validators.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open validators file: " + err.Error()})
			return
		}
		defer file.Close()
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read validators file: " + err.Error()})
			return
		}
		var validators []Validator
		if err := json.Unmarshal(bytes, &validators); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse validators data: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"count":      len(validators),
			"validators": validators,
		})
	}
}

// GetValidatorByIndex returns a single validator by index
func GetValidatorByIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		index := c.Param("index")
		file, err := os.Open("internal/data/validators.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open validators file: " + err.Error()})
			return
		}
		defer file.Close()
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read validators file: " + err.Error()})
			return
		}
		var validators []Validator
		if err := json.Unmarshal(bytes, &validators); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse validators data: " + err.Error()})
			return
		}
		for _, v := range validators {
			if v.Index == parseInt(index) {
				c.JSON(http.StatusOK, v)
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Validator not found"})
	}
}

func parseInt(s string) int {
	var i int
	_, err := fmt.Sscan(s, &i)
	if err != nil {
		return -1
	}
	return i
}
