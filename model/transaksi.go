package model

import "time"

type Transaksi struct {
	TransaksiID string    `json:"transaksi_id"`
	UserID      string    `json:"user_id"`
	TiketID     string    `json:"tiket_id"`
	Qty         int       `json:"qty"`
	Harga       int       `json:"harga"`
	QRCode      string    `json:"qrcode"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}
