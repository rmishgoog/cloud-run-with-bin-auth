package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Car struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Stock  int     `json:"stock"`
	Price  float64 `json:"price"`
	Origin string  `json:"origin"`
	Engine string  `json:"engine"`
}

var Cars = []Car{
	{ID: "1", Name: "Ford Mustang GT 5.0", Stock: 100, Price: 55000.99, Origin: "USA", Engine: "v8"},
	{ID: "2", Name: "Ford Explorer ST", Stock: 55, Price: 60000.99, Origin: "USA", Engine: "v6"},
	{ID: "4", Name: "Ford Escape Compact", Stock: 10, Price: 35000.89, Origin: "USA", Engine: "v4"},
	{ID: "5", Name: "Ford Edge LX Luxury", Stock: 1000, Price: 38000.89, Origin: "Germany", Engine: "v6"},
	{ID: "6", Name: "Ford Edge LX Base", Stock: 10, Price: 36000.89, Origin: "Germany", Engine: "v6"},
	{ID: "7", Name: "Ford Edge LX Base 2022", Stock: 10, Price: 36000.89, Origin: "Germany", Engine: "v6"},
	{ID: "8", Name: "Ford Edge LX Eco 2022", Stock: 10, Price: 54000.89, Origin: "Germany", Engine: "v6"},
}

func GetCarsHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, Cars)
}

func AddCarsHandler(c *gin.Context) {
	var newCar Car
	if err := c.BindJSON(&newCar); err != nil {
		return
	}
	Cars = append(Cars, newCar)
	c.IndentedJSON(http.StatusCreated, newCar)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/cars", GetCarsHandler)
	router.POST("/cars", AddCarsHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
		log.Printf("main.go: defaulting to port %s", port)
	}
	log.Printf("listening on port %s", port)
	router.Run(":" + port + "")
}
