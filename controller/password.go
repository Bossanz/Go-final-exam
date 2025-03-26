package controller

import (
	"go-final/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ChangePasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ฟังก์ชั่นสำหรับการเปลี่ยนรหัสผ่าน
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	var req ChangePasswordRequest

	// 1. ตรวจสอบรูปแบบ JSON
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "รูปแบบข้อมูลไม่ถูกต้อง",
			"details": err.Error(),
		})
		return
	}

	// 2. ค้นหาลูกค้าจาก email
	var customer model.Customer
	if err := c.DB.Where("email = ?", req.Email).First(&customer).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบลูกค้าด้วยอีเมลที่ให้มา"})
		return
	}

	// 3. ตรวจสอบรหัสผ่านเก่ากับรหัสที่เก็บในฐานข้อมูล
	err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(req.OldPassword))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "รหัสผ่านเก่าไม่ถูกต้อง",
		})
		return
	}

	// 4. แฮชรหัสผ่านใหม่
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "ไม่สามารถแฮชรหัสผ่านใหม่ได้",
		})
		return
	}

	// 5. อัพเดตรหัสผ่านใหม่ในฐานข้อมูล
	customer.Password = string(hashedPassword)
	if err := c.DB.Save(&customer).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "ไม่สามารถอัพเดตข้อมูลรหัสผ่านได้",
		})
		return
	}

	// 6. ส่งข้อความการเปลี่ยนรหัสผ่านสำเร็จ
	ctx.JSON(http.StatusOK, gin.H{
		"message": "รหัสผ่านถูกเปลี่ยนเรียบร้อย",
	})
}

// ฟังก์ชั่นตั้งค่าเส้นทาง (routes) ของ API
func SetupPasswordRoutes(router *gin.Engine, db *gorm.DB) {
	authController := NewAuthController(db)

	api := router.Group("/api")
	{
		// เส้นทางการเปลี่ยนรหัสผ่าน
		api.POST("/auth/change-password", authController.ChangePassword)
	}
}
