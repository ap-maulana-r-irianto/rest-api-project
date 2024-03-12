package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

var db *gorm.DB
var err error

type Order struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CustomerName string    `gorm:"column:customer_name" json:"customerName"`
	OrderedAt    time.Time `gorm:"column:ordered_at" json:"orderedAt"`
	Items        []Item    `gorm:"foreignKey:OrderID" json:"items"`
}

type Item struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Code        string `gorm:"column:code" json:"itemCode"`
	Description string `gorm:"column:description" json:"description"`
	Quantity    int    `gorm:"column:quantity" json:"quantity"`
	OrderID     uint   `gorm:"column:order_id"`
}

func main() {
	dsn := "host=localhost user=postgres dbname=maulanarafaelirianto password= sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Migrate the schema
	db.AutoMigrate(&Order{}, &Item{})

	// Set up Gin
	r := gin.Default()

	// Register your endpoints here
	r.POST("/orders", CreateOrder)
	r.GET("/orders", GetOrders)
	r.GET("/orders/:id", GetOrderById)
	r.PUT("/orders/:id", UpdateOrder)
	r.DELETE("/orders/:id", DeleteOrder)

	r.Run() // listen and serve on 0.0.0.0:8080 (default)
}

func CreateOrder(c *gin.Context) {
	var order Order

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Adding order with items
	if result := db.Create(&order); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func GetOrders(c *gin.Context) {
	var orders []Order
	if result := db.Preload("Items").Find(&orders); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func GetOrderById(c *gin.Context) {
	var order Order
	id := c.Param("id")
	if result := db.Preload("Items").First(&order, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func UpdateOrder(c *gin.Context) {
	var order Order
	id := c.Param("id")
	if err := db.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&order).Updates(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "order": order})
}

func DeleteOrder(c *gin.Context) {
	id := c.Param("id")
	if err := db.Where("order_id = ?", id).Delete(&Item{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result := db.Where("id = ?", id).Delete(&Order{}); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success delete": true})
}
