package entity

// Channel entity describing subscription to the channels of Coinbase Websocket feed
type Channel struct {
	Name       string   `json:"name"`
	ProductIds []string `json:"product_ids" mapstructure:"product_ids"`
}

// Subscription response from coinbase with channels that was subscribed to
type Subscription struct {
	Base
	Type     string    `json:"type"`
	Channels []Channel `json:"channels"`
}

// SubscriptionRequest to send to coinbase
type SubscriptionRequest struct {
	Type       string   `json:"type"`
	Channels   []string `json:"channels"`
	ProductIds []string `json:"product_ids" mapstructure:"product_ids"`
}
