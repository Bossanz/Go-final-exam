package main

import (
	"fmt"
	"go-final/controller"
	"go-final/model"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// โหลดค่าคอนฟิกจากไฟล์ config
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	// ดึงค่า DSN จาก config
	dsn := viper.GetString("mysql.dsn")
	fmt.Println("Connecting to database:", dsn)

	// เชื่อมต่อฐานข้อมูล
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	fmt.Println("Connection successful")

	// AutoMigrate ให้แน่ใจว่าตาราง `Customer` ถูกต้อง
	db.AutoMigrate(&model.Customer{})

	// สร้าง router ของ Gin
	router := gin.Default()

	// ตั้งค่า API routes
	controller.SetupRoutes(router, db)
	controller.SetupPasswordRoutes(router, db)

	// เริ่มเซิร์ฟเวอร์
	router.Run(":8080")
}
