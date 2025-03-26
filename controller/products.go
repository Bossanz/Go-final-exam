package controller

import (
	"go-final/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductController struct {
	DB *gorm.DB
}

func NewProductController(db *gorm.DB) *ProductController {
	return &ProductController{DB: db}
}

// ฟังก์ชันค้นหาสินค้า
func (c *ProductController) SearchProducts(ctx *gin.Context) {
	var searchParams struct {
		Name     string  `json:"name"`
		MinPrice float64 `json:"min_price"`
		MaxPrice float64 `json:"max_price"`
	}

	// ดึงพารามิเตอร์จาก query
	if err := ctx.ShouldBindQuery(&searchParams); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search parameters"})
		return
	}

	// สร้าง query สำหรับค้นหาสินค้า
	query := c.DB.Model(&model.Product{})

	// ถ้ามีการระบุชื่อสินค้า, ใช้ LIKE สำหรับค้นหา
	if searchParams.Name != "" {
		// ปรับ query ให้ค้นหาชื่อสินค้าแบบเปรียบเทียบที่ถูกต้อง
		query = query.Where("product_name LIKE ?", "%"+searchParams.Name+"%")
	}

	// ถ้ามีการระบุราคาต่ำสุด, ใช้เงื่อนไขราคาต่ำสุด
	if searchParams.MinPrice > 0 {
		query = query.Where("price >= ?", searchParams.MinPrice)
	}

	// ถ้ามีการระบุราคาสูงสุด, ใช้เงื่อนไขราคาสูงสุด
	if searchParams.MaxPrice > 0 {
		query = query.Where("price <= ?", searchParams.MaxPrice)
	}

	// ค้นหาสินค้าตามเงื่อนไขที่กำหนด
	var products []model.Product
	result := query.Find(&products)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching products", "details": result.Error.Error()})
		return
	}

	// ถ้าพบสินค้า, ส่งกลับข้อมูล
	if len(products) > 0 {
		ctx.JSON(http.StatusOK, products)
	} else {
		// ถ้าไม่พบสินค้า, แจ้งว่าไม่พบสินค้าตามเงื่อนไขที่กำหนด
		ctx.JSON(http.StatusNotFound, gin.H{"message": "No products found matching the search criteria"})
	}
}

// เส้นทางการตั้งค่า API
func SetupProductRoutes(router *gin.Engine, db *gorm.DB) {
	productController := NewProductController(db)

	api := router.Group("/api")
	{
		// เส้นทางการค้นหาสินค้า
		api.GET("/products/search", productController.SearchProducts)
	}
}
