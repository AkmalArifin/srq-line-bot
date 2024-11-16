package main

import (
	"os"

	"example.com/yahfaz/db"
	"example.com/yahfaz/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		panic(err.Error())
	}

	// Initialize DB
	db.InitDB()

	r := gin.Default()

	routes.RegisterRoutes(r)

	port := ":" + os.Getenv("PORT")
	err = r.Run(port)

	if err != nil {
		panic(err.Error())
	}

}
