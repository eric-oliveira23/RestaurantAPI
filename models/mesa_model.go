package model

type Pedido struct {
	Item       string  `json:"item"`
	Quantidade int     `json:"quantidade"`
	Valor      float64 `json:"valor"`
	Hash       string  `json:"hash" bson:"hash"`
	Imagem     string  `json:"imagem"`
}

type Mesa struct {
	Numero  int      `json:"numero"`
	Pedidos []Pedido `json:"pedidos"`
	Total   float64  `json:"total"`
}
