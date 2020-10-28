package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbport := os.Getenv("DB_PORT")
	dbuser := os.Getenv("DB_USER")
	dbpass := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_PASS")
	serverport := os.Getenv("SERVER_PORT")

	dbstr := fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbuser, dbpass, dbport, dbname)

	db, err := gorm.Open("mysql", dbstr)
	if err != nil {
		log.Fatalf("Got error when connect database, the error is '%v'", err)
	}

	r := gin.Default()
	r.Use(SetDBtoContext(db))

	r.POST("/model", Create)
	r.GET("/model", Read)
	r.PUT("/model/:id", Update)
	r.DELETE("/model/:id", Delete)

	r.Run(":8881" + serverport)
}

func Create(c *gin.Context) {
	c.JSON(200, "create")
}

func Read(c *gin.Context) {
	db := DBInstance(c)
	id := c.Params.ByName("id")

	type B struct {
		Brand string `json:"brand"`
		Nbr   int    `json:"nbr"`
	}
	var brand []B
	query := "SELECT * FROM DB_NAME.TABLE_NAME where id = ? limit 1;"
	db.Raw(query, id).Scan(&brand)
	c.JSON(200, brand)
}

func Update(c *gin.Context) {
	c.JSON(200, "update")
}

func Delete(c *gin.Context) {
	c.JSON(200, "delete")
}

func SetDBtoContext(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}

func DBInstance(c *gin.Context) *gorm.DB {
	return c.MustGet("DB").(*gorm.DB)
}
