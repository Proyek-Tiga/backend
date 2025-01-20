package model

import "time"

type Konser struct {
	KonserID      string    `json:"konser_id"`
	UserID        string    `json:"user_id"`
	LokasiID      string    `json:"lokasi_id"`
	NamaKonser    string    `json:"nama_konser"`
	TanggalKonser string    `json:"tanggal_konser"`
	JumlahTiket   int       `json:"jumlah_tiket"`
	Harga         int       `json:"harga"`
	Image         string    `json:"image"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Menambahkan nama user dan lokasi
	UserName   string `json:"user_name"`
	LokasiName string `json:"lokasi_name"`
}
