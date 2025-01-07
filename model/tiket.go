package model

import "time"

type Tiket struct {
	TiketID   string    `json:"tiket_id"`
	KonserID  string    `json:"konser_id"`
	NamaTiket string    `json:"nama_tiket"`
	Harga     int       `json:"harga"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
