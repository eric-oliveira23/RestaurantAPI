package routes

import (
	"endpoint/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.POST("/mesa/:numero/adicionar-item", handlers.AdicionarItem)
	router.GET("/mesas", handlers.ListarMesas)
	router.GET("/mesa/:numero", handlers.ObterDetalhesMesa)
}
