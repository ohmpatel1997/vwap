package entity

type Match struct {
	Base
	TradeID      int64  `json:"trade_id" mapstructure:"trade_id"`
	Sequence     int64  `json:"sequence"`
	MakerOrderID string `json:"maker_order_id" mapstructure:"maker_order_id"`
	TakerOrderID string `json:"taker_order_id" mapstructure:"taker_order_id"`
	Time         string `json:"time"`
	ProductID    string `json:"product_id" mapstructure:"product_id"`
	Size         string `json:"size"`
	Price        string `json:"price"`
	Side         string `json:"side"`
}
