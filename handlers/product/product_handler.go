package product_handler

import (
	"context"
	"encoding/base64"
	model "endpoint/models"
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

	var novoPedido model.Pedido
	if err := c.ShouldBindJSON(&novoPedido); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if novoPedido.Imagem == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhuma imagem enviada"})
		return
	}

	// Verifique se a imagem é base64 válida
	imagemBase64 := novoPedido.Imagem
	_, err := base64.StdEncoding.DecodeString(imagemBase64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "String base64 inválida"})
		return
	}

	novoPedido.Imagem = imagemBase64

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

func DeleteItem(c *gin.Context) {

	var item model.Pedido
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var filterItem = bson.M{"hash": item.Hash}

	var update = bson.M{
		"$pull": bson.M{"produtos": filterItem},
	}

	_, err := collection.UpdateOne(context.TODO(), filterItem, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removido com sucesso"})
}
