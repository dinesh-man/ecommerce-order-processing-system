package models

// Product represents an e-commerce product listing
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
	Brand       string  `json:"brand"`
	Rating      float64 `json:"rating"`
}
