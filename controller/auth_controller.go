package controller

import (
	"go-final/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	CustomerID  int    `json:"customer_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
}

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

// ฟังก์ชั่น Login ที่ตรวจสอบ email และ password ของผู้ใช้
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	// 1. ตรวจสอบรูปแบบ JSON
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "รูปแบบข้อมูลไม่ถูกต้อง",
			"details": err.Error(),
		})
		return
	}

	// 2. ค้นหาลูกค้าด้วย email
	var customer model.Customer
	result := c.DB.Where("email = ?", req.Email).First(&customer)
	if result.Error != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "อีเมลหรือรหัสผ่านไม่ถูกต้อง",
		})
		return
	}

	// 3. ตรวจสอบรหัสผ่านที่แฮชในฐานข้อมูล
	err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(req.Password))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "อีเมลหรือรหัสผ่านไม่ถูกต้อง",
		})
		return
	}

	// 4. เตรียม response โดยไม่ส่ง password
	response := LoginResponse{
		CustomerID:  customer.CustomerID,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Email:       customer.Email,
		PhoneNumber: customer.PhoneNumber,
		Address:     customer.Address,
	}

	// 5. ส่งข้อมูลกลับ
	ctx.JSON(http.StatusOK, response)
}

// ฟังก์ชั่นตั้งค่าเส้นทางของ API
func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	authController := NewAuthController(db)

	api := router.Group("/api")
	{
		api.POST("/auth/login", authController.Login)
	}
}
