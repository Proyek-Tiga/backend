package model

import "time"

type RequestLokasi struct {
	RequestID  string    `json:"request_id"`
	UserID     string    `json:"user_id"`
	NamaLokasi string    `json:"nama_lokasi"`
	Status     string    `json:"status"`
	Kapasitas  string    `json:"kapasitas"`
	Tiket      []Tiket   `json:"tiket"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
