package main

import (
	"endpoint/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	routes.SetupRoutes(router)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
