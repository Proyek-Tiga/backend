package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"project-tiket/config"
	"project-tiket/model"

	"github.com/gorilla/mux"
)

func AddTiket(w http.ResponseWriter, r *http.Request) {
	var tiket model.Tiket
	if err := json.NewDecoder(r.Body).Decode(&tiket); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Cek apakah jumlah tiket melebihi kapasitas jumlah_tiket pada konser
	var jumlahTiket, totalTiket int

	// Ambil jumlah_tiket dari konser
	queryJumlahTiket := `SELECT jumlah_tiket FROM konser WHERE konser_id = $1`
	err := config.DB.QueryRow(queryJumlahTiket, tiket.KonserID).Scan(&jumlahTiket)
	if err != nil {
		http.Error(w, "Gagal mendapatkan informasi jumlah_tiket konser: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Hitung total tiket yang sudah ada untuk konser ini
	queryTotalTiket := `SELECT COALESCE(SUM(jumlah_tiket), 0) FROM tiket WHERE konser_id = $1`
	err = config.DB.QueryRow(queryTotalTiket, tiket.KonserID).Scan(&totalTiket)
	if err != nil {
		http.Error(w, "Gagal menghitung total tiket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Periksa apakah total tiket ditambah tiket baru melebihi kapasitas
	if totalTiket+tiket.JumlahTiket > jumlahTiket {
		http.Error(w, "Jumlah tiket melebihi kapasitas konser", http.StatusBadRequest)
		return
	}

	// Insert tiket baru ke dalam database
	queryInsert := `
	  INSERT INTO tiket (konser_id, nama_tiket, harga, jumlah_tiket, created_at, updated_at)
	  VALUES ($1, $2, $3, $4, NOW(), NOW())
	  RETURNING tiket_id`

	var id string
	err = config.DB.QueryRow(queryInsert, tiket.KonserID, tiket.NamaTiket, tiket.Harga, tiket.JumlahTiket).Scan(&id)
	if err != nil {
		http.Error(w, "Gagal menambahkan tiket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kirim response sukses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Tiket added successfully",
		"id":      id,
	})
}

func GetTiket(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(`
	  SELECT
		t.tiket_id,
		t.konser_id,
		t.nama_tiket,
		t.jumlah_tiket,
		t.harga,
		t.created_at,
		t.updated_at,
		k.nama_konser
	  FROM
		tiket t
	  JOIN
		konser k ON t.konser_id = k.konser_id
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TiketResponse struct {
		TiketID     string  `json:"tiket_id"`
		KonserID    string  `json:"konser_id"`
		NamaTiket   string  `json:"nama_tiket"`
		JumlahTiket int     `json:"jumlah_tiket"`
		Harga       float64 `json:"harga"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
		NamaKonser  string  `json:"nama_konser"`
	}

	var tiketData []TiketResponse

	for rows.Next() {
		var tiket TiketResponse
		err := rows.Scan(
			&tiket.TiketID,
			&tiket.KonserID,
			&tiket.NamaTiket,
			&tiket.JumlahTiket,
			&tiket.Harga,
			&tiket.CreatedAt,
			&tiket.UpdatedAt,
			&tiket.NamaKonser,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tiketData = append(tiketData, tiket)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tiketData)
}

func GetTiketByID(w http.ResponseWriter, r *http.Request) {
	// Mengambil tiket_id dari URL parameter
	tiketID := mux.Vars(r)["id"]

	// Menjalankan query untuk mengambil tiket berdasarkan tiket_id
	row := config.DB.QueryRow(`
	  SELECT
		t.tiket_id,
		t.konser_id,
		t.nama_tiket,
		t.jumlah_tiket,
		t.harga,
		t.created_at,
		t.updated_at,
		k.nama_konser
	  FROM
		tiket t
	  JOIN
		konser k ON t.konser_id = k.konser_id
	  WHERE
		t.tiket_id = $1
	`, tiketID)

	// Menyiapkan variabel untuk menyimpan hasil query
	var tiket model.Tiket
	var konser model.Konser

	// Melakukan pemindaian hasil query ke dalam variabel tiket dan konser
	err := row.Scan(
		&tiket.TiketID,
		&tiket.KonserID,
		&tiket.NamaTiket,
		&tiket.JumlahTiket,
		&tiket.Harga,
		&tiket.CreatedAt,
		&tiket.UpdatedAt,
		&konser.NamaKonser,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Tiket tidak ditemukan", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Membuat response untuk tiket dengan hanya menampilkan nama konser
	response := struct {
		model.Tiket
		Konser string `json:"konser"`
	}{
		Tiket:  tiket,
		Konser: konser.NamaKonser,
	}

	// Mengirimkan response dalam format JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateTiket(w http.ResponseWriter, r *http.Request) {
	// Ambil tiket_id dari parameter URL
	vars := mux.Vars(r)
	tiketID := vars["id"]
	if tiketID == "" {
		http.Error(w, "Parameter tiket_id is required", http.StatusBadRequest)
		return
	}

	// Decode input JSON ke struct
	var tiket model.Tiket
	if err := json.NewDecoder(r.Body).Decode(&tiket); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Cek apakah tiket yang akan diupdate ada di database
	var oldJumlahTiket, konserID string
	queryCheckTiket := `SELECT konser_id, jumlah_tiket FROM tiket WHERE tiket_id = $1`
	err := config.DB.QueryRow(queryCheckTiket, tiketID).Scan(&konserID, &oldJumlahTiket)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Tiket not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve tiket: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Cek kapasitas tiket konser
	var jumlahTiket, totalTiket int

	// Ambil jumlah_tiket dari konser
	queryJumlahTiket := `SELECT jumlah_tiket FROM konser WHERE konser_id = $1`
	err = config.DB.QueryRow(queryJumlahTiket, konserID).Scan(&jumlahTiket)
	if err != nil {
		http.Error(w, "Gagal mendapatkan informasi jumlah_tiket konser: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Hitung total tiket tanpa memperhitungkan tiket yang akan diupdate
	queryTotalTiket := `SELECT COALESCE(SUM(jumlah_tiket), 0) FROM tiket WHERE konser_id = $1 AND tiket_id != $2`
	err = config.DB.QueryRow(queryTotalTiket, konserID, tiketID).Scan(&totalTiket)
	if err != nil {
		http.Error(w, "Gagal menghitung total tiket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Periksa apakah total tiket ditambah tiket baru melebihi kapasitas
	if totalTiket+tiket.JumlahTiket > jumlahTiket {
		http.Error(w, "Jumlah tiket melebihi kapasitas konser", http.StatusBadRequest)
		return
	}

	// Update tiket di database
	queryUpdate := `
	  UPDATE tiket
	  SET nama_tiket = $1, harga = $2, jumlah_tiket = $3, updated_at = NOW()
	  WHERE tiket_id = $4
	`
	result, err := config.DB.Exec(queryUpdate, tiket.NamaTiket, tiket.Harga, tiket.JumlahTiket, tiketID)
	if err != nil {
		http.Error(w, "Failed to update tiket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No rows updated. Please check the tiket_id.", http.StatusBadRequest)
		return
	}

	// Kirim response sukses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Tiket updated successfully",
		"id":      tiketID,
	})
}

func DeleteTiket(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}

	query := `
	  DELETE FROM tiket
	  WHERE tiket_id=$1`

	// Execute the SQL statement
	result, err := config.DB.Exec(query, idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "No rows were deleted", http.StatusNotFound)
		return
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "tiket deleted successfully",
	})
}

func GetTiketByKonser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["konser_id"]
	if !ok {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}

	// Query database untuk mendapatkan data tiket berdasarkan UUID konser
	var tikets model.Tiket
	err := config.DB.QueryRow(
		"SELECT tiket_id, konser_id, nama_tiket, jumlah_tiket, harga, created_at, updated_at FROM tiket WHERE konser_id = $1",
		id,
	).Scan(&tikets.TiketID, &tikets.KonserID, &tikets.NamaTiket, &tikets.JumlahTiket, &tikets.Harga, &tikets.CreatedAt, &tikets.UpdatedAt)
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Berikan respons dalam format JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tikets); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}
