package model

import "time"

// Payment represents a payment transaction.
type Payment struct {
    PaymentID   string    `json:"payment_id" bson:"payment_id"`
    OrderID     string    `json:"order_id" bson:"order_id"`
    UserID      string    `json:"user_id" bson:"user_id"`
    GrossAmount float64   `json:"gross_amount" bson:"gross_amount"`
    SnapURL     string    `json:"snap_url" bson:"snap_url"`
    Status      string    `json:"status" bson:"status"`
    CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}
