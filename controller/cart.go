package controller

import (
	"go-final/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CartController struct {
	DB *gorm.DB
}

func NewCartController(db *gorm.DB) *CartController {
	return &CartController{DB: db}
}

// ฟังก์ชันเพิ่มสินค้าลงในรถเข็น
func (c *CartController) AddToCart(ctx *gin.Context) {
	var req struct {
		CustomerID uint   `json:"customer_id" binding:"required"`
		ProductID  uint   `json:"product_id" binding:"required"`
		Quantity   uint   `json:"quantity" binding:"required"`
		CartName   string `json:"cart_name" binding:"required"`
	}

	// รับข้อมูลจาก request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "รูปแบบข้อมูลไม่ถูกต้อง",
			"details": err.Error(),
		})
		return
	}

	// ค้นหาลูกค้าจาก CustomerID
	var customer model.Customer
	if err := c.DB.Where("customer_id = ?", req.CustomerID).First(&customer).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบลูกค้าด้วย ID ที่ให้มา"})
		return
	}

	// ค้นหารถเข็นที่ลูกค้าต้องการใช้งาน
	var cart model.Cart
	result := c.DB.Where("customer_id = ? AND cart_name = ?", customer.CustomerID, req.CartName).First(&cart)

	// ถ้ารถเข็นไม่มีให้สร้างใหม่
	if result.Error != nil {
		cart = model.Cart{
			CustomerID: customer.CustomerID,
			CartName:   req.CartName,
		}
		if err := c.DB.Create(&cart).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้างรถเข็นได้"})
			return
		}
	}

	// ค้นหาสินค้าที่ลูกค้าต้องการเพิ่ม
	var product model.Product
	if err := c.DB.Where("product_id = ?", req.ProductID).First(&product).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสินค้านี้ในฐานข้อมูล"})
		return
	}

	// ค้นหาว่าสินค้านี้มีอยู่ในรถเข็นแล้วหรือไม่
	var cartItem model.CartItem
	result = c.DB.Where("cart_id = ? AND product_id = ?", cart.CartID, req.ProductID).First(&cartItem)

	if result.Error == nil {
		// ถ้ามีสินค้าในรถเข็นแล้วให้เพิ่มจำนวน
		cartItem.Quantity += int(req.Quantity)
		if err := c.DB.Save(&cartItem).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัพเดตจำนวนสินค้าในรถเข็นได้"})
			return
		}
	} else {
		// ถ้าไม่มีสินค้าในรถเข็นให้เพิ่มสินค้าลงไป
		cartItem = model.CartItem{
			CartID:    cart.CartID,
			ProductID: int(req.ProductID),
			Quantity:  int(req.Quantity),
		}
		if err := c.DB.Create(&cartItem).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเพิ่มสินค้าในรถเข็นได้"})
			return
		}
	}

	// ส่งข้อความการเพิ่มสินค้าสำเร็จ
	ctx.JSON(http.StatusOK, gin.H{
		"message": "เพิ่มสินค้าลงในรถเข็นเรียบร้อย",
	})
}

// ฟังก์ชันดูสินค้าที่อยู่ในรถเข็น
func (c *CartController) ViewCart(ctx *gin.Context) {
	var req struct {
		CustomerID uint   `json:"customer_id" binding:"required"`
		CartName   string `json:"cart_name" binding:"required"`
	}

	// รับข้อมูลจาก request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "รูปแบบข้อมูลไม่ถูกต้อง",
			"details": err.Error(),
		})
		return
	}

	// ค้นหาลูกค้าจาก CustomerID
	var customer model.Customer
	if err := c.DB.Where("customer_id = ?", req.CustomerID).First(&customer).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบลูกค้าด้วย ID ที่ให้มา"})
		return
	}

	// ค้นหารถเข็นของลูกค้า
	var cart model.Cart
	if err := c.DB.Where("customer_id = ? AND cart_name = ?", customer.CustomerID, req.CartName).First(&cart).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรถเข็นนี้"})
		return
	}

	// ค้นหารายการสินค้าทั้งหมดในรถเข็น
	var cartItems []model.CartItem
	if err := c.DB.Where("cart_id = ?", cart.CartID).Find(&cartItems).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลสินค้าจากรถเข็นได้"})
		return
	}

	// เตรียมข้อมูลสินค้าที่จะส่งกลับ
	var productsInCart []gin.H
	for _, cartItem := range cartItems {
		var product model.Product
		if err := c.DB.Where("product_id = ?", cartItem.ProductID).First(&product).Error; err == nil {
			productsInCart = append(productsInCart, gin.H{
				"product_id": product.ProductID,
				"name":       product.ProductName,
				"price":      product.Price,
				"quantity":   cartItem.Quantity,
			})
		}
	}

	// ส่งข้อมูลสินค้าทั้งหมดในรถเข็นกลับ
	ctx.JSON(http.StatusOK, gin.H{
		"cart_name": req.CartName,
		"products":  productsInCart,
	})
}

// เส้นทางการตั้งค่า API
func SetupCartRoutes(router *gin.Engine, db *gorm.DB) {
	cartController := NewCartController(db)

	api := router.Group("/api")
	{
		// เส้นทางการเพิ่มสินค้าในรถเข็น
		api.POST("/cart/add", cartController.AddToCart)

		// เส้นทางการดูสินค้าจากรถเข็น
		api.POST("/cart/view", cartController.ViewCart)
	}
}
