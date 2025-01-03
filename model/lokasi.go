package model

import "time"

type Lokasi struct {
  LokasiID string `json:"lokasi_id"`
  Lokasi   string `json:"lokasi"`
  Tiket    int    `json:"tiket"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}
