package product_handler

import (
	"context"
	model "endpoint/models"
	utils "endpoint/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}

	collection = client.Database("restaurant_db").Collection("products")
}

func AddProduct(c *gin.Context) {

	// Dados do novo pedido a ser adicionado (a partir do corpo da solicitação JSON)
	var novoPedido model.Pedido
	if err := c.ShouldBindJSON(&novoPedido); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	randomHash, err := utils.HashGenerator(6)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	novoPedido.Hash = randomHash

	fmt.Println("Pedido a ser inserido:", novoPedido)

	_, err = collection.InsertOne(context.TODO(), novoPedido)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Produto cadastrado com sucesso"})
	}

}

func GetAllProducts(c *gin.Context) {
	// Consulte todas as mesas no banco de dados
	var products []model.Pedido
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var product model.Pedido
		err := cursor.Decode(&product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, product)
	}

	c.JSON(http.StatusOK, products)

}
