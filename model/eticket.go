package model

type ETiket struct {
	UserName        string `json:"user_name"`        // Nama user yang membeli tiket
	TiketName       string `json:"tiket_name"`       // Nama tiket
	KonserName      string `json:"konser_name"`      // Nama konser
	KonserLocation  string `json:"konser_location"`  // Lokasi konser
	QRCode          string `json:"qr_code"`          // QR Code tiket
	TransaksiStatus string `json:"transaksi_status"` // Status transaksi (settlement, pending, dsb.)
	TanggalKonser   string `json:"tanggal_konser"`   // Tanggal konser
}
