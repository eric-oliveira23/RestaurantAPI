package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Pedido struct {
	Item       string `json:"item"`
	Quantidade int    `json:"quantidade"`
}

type Mesa struct {
	Numero  int      `json:"numero"`
	Pedidos []Pedido `json:"pedidos"`
}

func main() {
	// Configuração do servidor Gin
	router := gin.Default()

	// Configuração da conexão com o MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("restaurant_db").Collection("restaurant")

	// Rota para adicionar um item a uma mesa
	router.POST("/mesa/:numero/adicionar-item", func(c *gin.Context) {
		// Obter o número da mesa dos parâmetros da URL
		numeroMesa := c.Param("numero")

		// Converter o número da mesa para int
		numeroMesaInt, err := strconv.Atoi(numeroMesa)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Número de mesa inválido"})
			return
		}

		// Dados do novo pedido a ser adicionado (a partir do corpo da solicitação JSON)
		var novoPedido Pedido
		if err := c.ShouldBindJSON(&novoPedido); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verificar se a mesa com o número especificado já existe
		filter := bson.M{"numero": numeroMesaInt}
		var mesa Mesa
		err = collection.FindOne(context.TODO(), filter).Decode(&mesa)
		if err == mongo.ErrNoDocuments {
			// A mesa não existe, então criamos uma nova mesa com o pedido
			novaMesa := Mesa{
				Numero:  numeroMesaInt,
				Pedidos: []Pedido{novoPedido},
			}
			_, err := collection.InsertOne(context.TODO(), novaMesa)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			// A mesa existe, então adicionamos o pedido à lista de pedidos da mesa
			update := bson.M{"$push": bson.M{"pedidos": novoPedido}}
			_, err := collection.UpdateOne(context.TODO(), filter, update)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Pedido adicionado com sucesso"})
	})

	// Todas as mesas
	router.GET("/mesas", func(c *gin.Context) {
		// Consulte todas as mesas no banco de dados
		var mesas []Mesa
		cursor, err := collection.Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(context.TODO())

		for cursor.Next(context.TODO()) {
			var mesa Mesa
			err := cursor.Decode(&mesa)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			mesas = append(mesas, mesa)
		}

		c.JSON(http.StatusOK, mesas)
	})

	// Detalhes mesa
	router.GET("/mesa/:numero", func(c *gin.Context) {
		// Obter o número da mesa dos parâmetros da URL
		numeroMesa := c.Param("numero")

		// Converter o número da mesa para int
		numeroMesaInt, err := strconv.Atoi(numeroMesa)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Número de mesa inválido"})
			return
		}

		// Consulte os detalhes da mesa específica no banco de dados
		filter := bson.M{"numero": numeroMesaInt}
		var mesa Mesa
		err = collection.FindOne(context.TODO(), filter).Decode(&mesa)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Mesa não encontrada"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, mesa)
	})

	// Inicie o servidor Gin
	err = router.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
