package model

import "time"

type Konser struct {
	KonserID      string    `json:"konser_id"`
	UserID        string    `json:"user_id"`
	LokasiID      string    `json:"lokasi_id"`
	TiketID       string    `json:"tiket_id"`
	NamaKonser    string    `json:"nama_konser"`
	TanggalKonser time.Time `json:"tanggal_konser"`
	JumlahTiket   int       `json:"jumlah_tiket"`
	Harga         int       `json:"harga"`
	Image         string    `json:"image"`
	JenisBank     string    `json:"jenis_bank"`
	AtasNama      string    `json:"atas_nama"`
	Rekening      int       `json:"rekening"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
