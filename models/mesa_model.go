package models

type Pedido struct {
	Item       string `json:"item"`
	Quantidade int    `json:"quantidade"`
}

type Mesa struct {
	Numero  int      `json:"numero"`
	Pedidos []Pedido `json:"pedidos"`
}
