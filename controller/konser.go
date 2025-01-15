package controller

import (
	"encoding/json"
	"net/http"

	"project-tiket/config"
	"project-tiket/model"

	"github.com/gorilla/mux"
	// "github.com/gorilla/mux"
)

func AddKonser(w http.ResponseWriter, r *http.Request) {
	var konser model.Konser
	if err := json.NewDecoder(r.Body).Decode(&konser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ambil jumlah tiket dari lokasi yang dipilih
	var tiketLokasi int
	queryLokasi := `SELECT tiket FROM lokasi WHERE lokasi_id = $1`
	err := config.DB.QueryRow(queryLokasi, konser.LokasiID).Scan(&tiketLokasi)
	if err != nil {
		http.Error(w, "Failed to fetch ticket count for the location: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Tetapkan jumlah tiket konser sesuai dengan tiket di lokasi
	konser.JumlahTiket = tiketLokasi

	// Tetapkan status konser menjadi "pending"
	konser.Status = "pending"

	// Query untuk menambahkan konser
	query := `
	INSERT INTO konser (user_id, lokasi_id, tiket_id, nama_konser, tanggal_konser, jumlah_tiket, harga, image, jenis_bank, atas_nama, rekening, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
	RETURNING konser_id`

	var id string
	err = config.DB.QueryRow(query, 
		konser.UserID, 
		konser.LokasiID,
		konser.TiketID,
		konser.NamaKonser,
		konser.TanggalKonser, 
		konser.JumlahTiket, 
		konser.Harga, 
		konser.Image, 
		konser.JenisBank, 
		konser.AtasNama, 
		konser.Rekening, 
		konser.Status, 
	).Scan(&id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the newly created ID in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Konser added successfully",
		"id":      id,
	})
}


func UpdateKonser(w http.ResponseWriter, r *http.Request) {
	// Parse ID from the URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Decode the JSON request body
	var konser model.Konser
	if err := json.NewDecoder(r.Body).Decode(&konser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update query
	query := `
		UPDATE konser
		SET user_id = $1, 
			lokasi_id = $2,
			tiket_id = $3, 
			nama_konser = $4, 
			tanggal_konser = $5, 
			jumlah_tiket = $6, 
			harga = $7, 
			image = $8, 
			jenis_bank = $9, 
			atas_nama = $10, 
			rekening = $11, 
			status = $12, 
			updated_at = NOW()
		WHERE konser_id = $13`

	// Execute the query
	_, err := config.DB.Exec(query, 
		konser.UserID, 
		konser.LokasiID, 
		konser.TiketID,
		konser.NamaKonser, 
		konser.TanggalKonser, 
		konser.JumlahTiket, 
		konser.Harga, 
		konser.Image, 
		konser.JenisBank, 
		konser.AtasNama, 
		konser.Rekening, 
		konser.Status, 
		id,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Konser updated successfully",
		"id":      id,
	})
}

func GetAllKonser(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT konser_id, user_id, lokasi_id, tiket_id, nama_konser, tanggal_konser, jumlah_tiket, harga, image, jenis_bank, atas_nama, rekening, status, created_at, updated_at
		FROM konser`

	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var konserList []model.Konser
	for rows.Next() {
		var konser model.Konser
		if err := rows.Scan(
			&konser.KonserID,
			&konser.UserID,
			&konser.LokasiID,
			&konser.TiketID,
			&konser.NamaKonser,
			&konser.TanggalKonser,
			&konser.JumlahTiket,
			&konser.Harga,
			&konser.Image,
			&konser.JenisBank,
			&konser.AtasNama,
			&konser.Rekening,
			&konser.Status,
			&konser.CreatedAt,
			&konser.UpdatedAt,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		konserList = append(konserList, konser)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(konserList)
}

func GetKonserByID(w http.ResponseWriter, r *http.Request) {
	// Ambil parameter ID dari URL
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}

	// Query database untuk mendapatkan data konser berdasarkan ID
	var konser model.Konser
	err := config.DB.QueryRow(
		`SELECT konser_id, user_id, lokasi_id, tiket_id, nama_konser, tanggal_konser, jumlah_tiket, harga, image, jenis_bank, atas_nama, rekening, status, created_at, updated_at 
		FROM konser 
		WHERE konser_id = $1`,
		id,
	).Scan(
		&konser.KonserID,
		&konser.UserID,
		&konser.LokasiID,
		&konser.TiketID,
		&konser.NamaKonser,
		&konser.TanggalKonser,
		&konser.JumlahTiket,
		&konser.Harga,
		&konser.Image,
		&konser.JenisBank,
		&konser.AtasNama,
		&konser.Rekening,
		&konser.Status,
		&konser.CreatedAt,
		&konser.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Berikan respons dalam format JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(konser); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

func UpdateKonserStatus(w http.ResponseWriter, r *http.Request) {
	// Ambil ID konser dari URL parameter
	konserID := mux.Vars(r)["id"]

	// Define struktur untuk status baru
	var status struct {
		Status string `json:"status"`
	}

	// Decode JSON body untuk mendapatkan status baru
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Pastikan status yang diberikan valid
	if status.Status != "pending" && status.Status != "approved" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	// Mulai transaksi untuk memastikan atomicity
	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update status konser
	query := `UPDATE konser SET status = $1 WHERE konser_id = $2 AND status = 'pending'`
	_, err = tx.Exec(query, status.Status, konserID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Jika status yang diberikan adalah "approved", lakukan penanganan lebih lanjut (misalnya menambah tiket atau tindakan lainnya)
	if status.Status == "approved" {
		// Anda bisa menambahkan logika lebih lanjut untuk "approved" di sini
		// Misalnya, menambahkan data ke tabel lain jika perlu
	}

	// Commit transaksi jika semuanya berjalan lancar
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Konser status updated successfully"})
}

// GetApprovedConcerts menampilkan konser dengan status "approved"
func GetApprovedConcerts(w http.ResponseWriter, r *http.Request) {
	var concerts []model.Konser

	// Query untuk mengambil data konser dengan status "approved"
	query := `
			SELECT * FROM konser WHERE status = $1`

	rows, err := config.DB.Query(query, "approved")
	if err != nil {
		http.Error(w, "Failed to fetch approved concerts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var concert model.Konser
		err := rows.Scan(
			&concert.KonserID,
			&concert.UserID,
			&concert.LokasiID,
			&concert.NamaKonser,
			&concert.TanggalKonser,
			&concert.JumlahTiket,
			&concert.Harga,
			&concert.Image,
			&concert.JenisBank,
			&concert.AtasNama,
			&concert.Rekening,
			&concert.Status,
			&concert.CreatedAt,
			&concert.UpdatedAt,
			&concert.TiketID,
		)
		if err != nil {
			http.Error(w, "Failed to scan concert data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		concerts = append(concerts, concert)
	}

	if len(concerts) == 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]model.Konser{})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(concerts)
}