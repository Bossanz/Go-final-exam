package controller

import (
	"go-final/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *gorm.DB
}

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

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

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

	// 3. ถ้ารหัสผ่านในฐานข้อมูลยังไม่ได้แฮช
	if err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(customer.Password)); err != nil {
		// แฮชรหัสผ่านใหม่
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "ไม่สามารถแฮชรหัสผ่านได้",
			})
			return
		}
		// อัพเดตฐานข้อมูลด้วยรหัสผ่านที่แฮชแล้ว
		customer.Password = string(hashedPassword)
		if err := c.DB.Save(&customer).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "ไม่สามารถอัพเดตข้อมูลรหัสผ่านได้",
			})
			return
		}
	}

	// 4. ตรวจสอบรหัสผ่านที่แฮชแล้ว
	err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(req.Password))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "อีเมลหรือรหัสผ่านไม่ถูกต้อง",
		})
		return
	}

	// 5. เตรียม response โดยไม่ส่ง password
	response := LoginResponse{
		CustomerID:  customer.CustomerID,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Email:       customer.Email,
		PhoneNumber: customer.PhoneNumber,
		Address:     customer.Address,
	}

	// 6. ส่งข้อมูลกลับ
	ctx.JSON(http.StatusOK, response)
}

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	authController := NewAuthController(db)

	api := router.Group("/api")
	{
		api.POST("/auth/login", authController.Login)
	}
}
