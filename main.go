package main

import (
	"example.com/yahfaz/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		panic(err.Error())
	}

	r := gin.Default()

	routes.RegisterRoutes(r)

	port := ":5050"

	r.Run(port)
}
