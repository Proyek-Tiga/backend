package model

import "time"

// Transaksi mewakili data transaksi yang terlibat dalam query
type TransaksiPenyelenggara struct {
	TransaksiID       string    `json:"transaksi_id"`
	TiketID           string    `json:"tiket_id"`
	PenyelenggaraID   string    `json:"penyelenggara_id"`
	PenyelenggaraName string    `json:"penyelenggara_name"`
	KonserName        string    `json:"konser_name"`
	PembeliID         string    `json:"pembeli_id"`
	PembeliName       string    `json:"pembeli_name"`
	TransaksiStatus   string    `json:"transaksi_status"`
	QRCode            string    `json:"qr_code"`
	TransaksiDate     time.Time `json:"transaksi_date"`
}
