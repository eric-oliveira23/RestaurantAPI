package mesa_handler

import (
	"context"
	model "endpoint/models"
	"net/http"
	"strconv"
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

	collection = client.Database("restaurant_db").Collection("tables")
}

// @Summary Adicionar item à mesa
// @Description Adiciona um item à mesa especificada
// @Tags Tables
// @Accept json
// @Produce json
// @Param numero path int true "Número da mesa"
// @Param item body model.Pedido true "Item a ser adicionado"
// @Success 200 {object} model.Pedido
// @Router /mesa/{numero}/adicionar-item [post]
func AddItem(c *gin.Context) {
	// Obter o número da mesa dos parâmetros da URL
	numeroMesa := c.Param("numero")

	// Converter o número da mesa para int
	numeroMesaInt, err := strconv.Atoi(numeroMesa)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Número de mesa inválido"})
		return
	}

	// Dados do novo pedido a ser adicionado (a partir do corpo da solicitação JSON)
	var novoPedido model.Pedido
	if err := c.ShouldBindJSON(&novoPedido); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Usar diretamente o valor da estrutura novoPedido.Valor
	valorFloat := novoPedido.Valor

	// Verificar se a mesa com o número especificado já existe
	filter := bson.M{"numero": numeroMesaInt}
	var mesa model.Mesa
	err = collection.FindOne(context.TODO(), filter).Decode(&mesa)
	if err == mongo.ErrNoDocuments {
		// A mesa não existe, então criamos uma nova mesa com o pedido
		novaMesa := model.Mesa{
			Numero:  numeroMesaInt,
			Pedidos: []model.Pedido{novoPedido},
			Total:   float64(valorFloat),
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
		update := bson.M{
			"$push": bson.M{"pedidos": novoPedido},
			"$inc":  bson.M{"total": float32(novoPedido.Valor)},
		}
		_, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pedido adicionado com sucesso"})

}

func ListTables(c *gin.Context) {
	// Consulte todas as mesas no banco de dados
	var mesas []model.Mesa
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var mesa model.Mesa
		err := cursor.Decode(&mesa)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		mesas = append(mesas, mesa)
	}

	c.JSON(http.StatusOK, mesas)
}

func GetTableDetails(c *gin.Context) {
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
	var mesa model.Mesa
	err = collection.FindOne(context.TODO(), filter).Decode(&mesa)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mesa não encontrada"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mesa)
}

func RemoveItems(c *gin.Context) {
	numeroMesa := c.Param("numero")

	// Converter o número da mesa para int
	numeroMesaInt, err := strconv.Atoi(numeroMesa)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Número de mesa inválido"})
		return
	}

	// Dados da lista de itens a serem removidos (a partir do corpo da solicitação JSON)
	var itensRemover []model.Pedido
	if err := c.ShouldBindJSON(&itensRemover); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se a mesa com o número especificado existe
	filter := bson.M{"numero": numeroMesaInt}
	var mesa model.Mesa
	err = collection.FindOne(context.TODO(), filter).Decode(&mesa)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mesa não encontrada"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Percorra a lista de itens a serem removidos e crie um filtro para encontrar os itens correspondentes
	for _, item := range itensRemover {
		filterItem := bson.M{"hash": item.Hash}
		// Remove o item da lista de pedidos da mesa
		update := bson.M{
			"$pull": bson.M{"pedidos": filterItem},
			"$inc":  bson.M{"total": -item.Valor},
		}

		// Remova a mesa se não houver mais pedidos nela
		if len(mesa.Pedidos) == 1 {
			_, err := collection.DeleteOne(context.TODO(), filter)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Mesa removida, pois não há mais pedidos"})
			return
		}

		_, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Itens removidos com sucesso"})
}
