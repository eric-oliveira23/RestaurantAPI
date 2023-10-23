package routes

import (
	product_handlers "endpoint/handlers/product"
	table_handlers "endpoint/handlers/table"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {

	// Tables
	router.POST("/mesa/:numero/adicionar-item", table_handlers.AddItem)
	router.GET("/mesas", table_handlers.ListTables)
	router.GET("/mesa/:numero", table_handlers.GetTableDetails)
	router.POST("/mesa/:numero/excluir", table_handlers.RemoveItems)

	// Products
	router.POST("/cadastrar-produto", product_handlers.AddProduct)
	router.GET("/produtos", product_handlers.GetAllProducts)
	router.POST("/remover-produto", product_handlers.DeleteItem)

}
