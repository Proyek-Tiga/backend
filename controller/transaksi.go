package controller

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"project-tiket/config"
	"project-tiket/model"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/skip2/go-qrcode"
)

func CreateTransaksi(w http.ResponseWriter, r *http.Request) {
	var transaksi model.Transaksi

	// Parse request body
	err := json.NewDecoder(r.Body).Decode(&transaksi)
	if err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	// Generate UUID untuk transaksi
	transaksi.TransaksiID = uuid.NewString()
	transaksi.CreatedAt = time.Now()
	transaksi.UpdatedAt = time.Now()

	// Insert ke database
	query := `
		INSERT INTO transaksi (transaksi_id, user_id, tiket_id, qty, harga, updated_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = config.DB.Exec(query, transaksi.TransaksiID, transaksi.UserID, transaksi.TiketID, transaksi.Qty, transaksi.Harga, transaksi.UpdatedAt, transaksi.CreatedAt)
	if err != nil {
		http.Error(w, "Gagal membuat transaksi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate QR code
	qrData := fmt.Sprintf("UserID: %s, TransaksiID: %s", transaksi.UserID, transaksi.TransaksiID)
	qrCode, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Gagal menghasilkan QR code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode QR code ke base64
	encodedQRCode := base64.StdEncoding.EncodeToString(qrCode)

	// Persiapkan data untuk upload ke GitHub
	githubToken := os.Getenv("GH_ACCESS_TOKEN")                                           // Token GitHub yang diambil dari environment
	repoOwner := "Proyek-Tiga"                                                            // Nama pemilik repository
	repoName := "images"                                                                  // Nama repository
	filePath := fmt.Sprintf("qrcode/%s_%d.png", transaksi.TransaksiID, time.Now().Unix()) // Lokasi file di repo GitHub
	uploadURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", repoOwner, repoName, filePath)

	// Payload untuk request GitHub API
	payload := map[string]string{
		"message": fmt.Sprintf("Tambah QR code untuk transaksi %s", transaksi.TransaksiID),
		"content": encodedQRCode,
	}
	payloadBytes, _ := json.Marshal(payload)

	// Membuat request ke GitHub API
	req, _ := http.NewRequest("PUT", uploadURL, bytes.NewReader(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Content-Type", "application/json")

	// Mengirimkan request ke GitHub API
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Gagal mengupload QR code ke GitHub: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Cek status code dari GitHub API
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Error dari GitHub API: %s", string(body)), http.StatusInternalServerError)
		return
	}

	// Ambil URL raw file dari response GitHub
	var githubResponse struct {
		Content struct {
			DownloadURL string `json:"download_url"`
		} `json:"content"`
	}
	err = json.NewDecoder(resp.Body).Decode(&githubResponse)
	if err != nil {
		http.Error(w, "Gagal mendekode response GitHub: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update URL QR code di database
	qrCodeURL := githubResponse.Content.DownloadURL
	updateQuery := `UPDATE transaksi SET qrcode = $1 WHERE transaksi_id = $2`
	_, err = config.DB.Exec(updateQuery, qrCodeURL, transaksi.TransaksiID)
	if err != nil {
		http.Error(w, "Gagal mengupdate transaksi dengan URL QR code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Mengirimkan response ke client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaksi)
}

func GetTransaksiByID(w http.ResponseWriter, r *http.Request) {
	// Parse transaksi_id from the URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Query to get the transaksi by ID
	query := `SELECT * FROM transaksi WHERE transaksi_id = $1`

	// Struct to hold the transaksi data
	var transaksi model.Transaksi

	// Execute the query
	err := config.DB.QueryRow(query, id).Scan(
		&transaksi.TransaksiID,
		&transaksi.UserID,
		&transaksi.TiketID,
		&transaksi.Qty,
		&transaksi.Harga,
		&transaksi.QRCode,
		&transaksi.CreatedAt,
		&transaksi.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Transaksi not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return the transaksi data as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaksi)
}
