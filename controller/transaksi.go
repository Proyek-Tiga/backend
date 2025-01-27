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
	"strings"
	"time"

	"project-tiket/config"
	"project-tiket/model"

	"github.com/dgrijalva/jwt-go"
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
	transaksi.Status = "pending"
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

func GetAllTransaksi(w http.ResponseWriter, r *http.Request) {
	// Query untuk mendapatkan data dari tabel transaksi dan payment
	rows, err := config.DB.Query(`
	  SELECT
		t.transaksi_id, t.user_id, t.tiket_id, t.qty, t.harga, t.qrcode, t.updated_at, t.created_at,
		COALESCE(p.payment_id, NULL) AS payment_id,
		COALESCE(p.order_id, NULL) AS order_id,
		COALESCE(p.user_id, NULL) AS user_id,
		COALESCE(p.gross_amount, 0) AS gross_amount,
		COALESCE(p.snap_url, '') AS snap_url,
		COALESCE(p.status, '') AS status,
		COALESCE(p.created_at, NULL) AS payment_created_at
	  FROM transaksi t
	  LEFT JOIN payment p ON t.transaksi_id = p.order_id
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Error mengambil data transaksi",
		})

		return
	}
	defer rows.Close()

	var result []map[string]interface{} // Menampung hasil sebagai map untuk fleksibilitas

	// Iterasi hasil query
	for rows.Next() {
		var transaksi model.Transaksi
		var payment model.Payment

		var paymentID sql.NullString
		var orderID sql.NullString
		var userID sql.NullString
		var paymentCreatedAt sql.NullTime

		err := rows.Scan(
			&transaksi.TransaksiID,
			&transaksi.UserID,
			&transaksi.TiketID,
			&transaksi.Qty,
			&transaksi.Harga,
			&transaksi.QRCode,
			&transaksi.UpdatedAt,
			&transaksi.CreatedAt,
			&paymentID, // Menggunakan sql.NullString
			&orderID,   // Menggunakan sql.NullString
			&userID,    // Menggunakan sql.NullString
			&payment.GrossAmount,
			&payment.SnapURL,
			&payment.Status,
			&paymentCreatedAt, // Menggunakan sql.NullTime
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Konversi NullString dan NullTime ke nilai asli
		if paymentID.Valid {
			payment.PaymentID = paymentID.String
		}
		if orderID.Valid {
			payment.OrderID = orderID.String
		}
		if userID.Valid {
			payment.UserID = userID.String
		}
		if paymentCreatedAt.Valid {
			payment.CreatedAt = paymentCreatedAt.Time
		}

		// Gabungkan data transaksi dan payment ke dalam map
		combined := map[string]interface{}{
			"transaksi": transaksi,
			"payment":   payment,
		}
		result = append(result, combined)
	}

	// Cek error pada iterasi rows
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Kirim response JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func GetETiket(w http.ResponseWriter, r *http.Request) {
	// Ambil Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		return
	}

	// Ekstrak token dari "Bearer <token>"
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Klaim untuk menyimpan informasi dari token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil // Sesuaikan `jwtKey` dengan secret key JWT Anda
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Ambil user_id dari klaim
	userID := claims.UserID // Pastikan klaim `UserID` sesuai dengan struktur Anda
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	// Query untuk mendapatkan data e-Tiket
	query := `
	  SELECT
		u.name AS user_name,
		k.nama_konser AS konser_name,
		tk.nama_tiket,
		k.tanggal_konser,
		l.lokasi AS konser_location,
		CASE
		  WHEN t.status = 'settlement' THEN t.qrcode
		  ELSE ''
		END AS qr_code,
		t.status AS transaksi_status
	  FROM transaksi t
	  JOIN users u ON t.user_id = u.user_id
	  JOIN tiket tk ON t.tiket_id = tk.tiket_id
	  JOIN konser k ON tk.konser_id = k.konser_id
	  JOIN lokasi l ON k.lokasi_id = l.lokasi_id
	  WHERE t.user_id = $1;
	`

	// Jalankan query
	rows, err := config.DB.Query(query, userID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parsing hasil query
	var tiketList []model.ETiket
	for rows.Next() {
		var tiket model.ETiket
		if err := rows.Scan(
			&tiket.UserName,
			&tiket.KonserName,
			&tiket.TiketName,
			&tiket.TanggalKonser,
			&tiket.KonserLocation,
			&tiket.QRCode,
			&tiket.TransaksiStatus,
		); err != nil {
			http.Error(w, "Failed to scan result: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tiketList = append(tiketList, tiket)
	}

	// Cek apakah ada data
	if len(tiketList) == 0 {
		http.Error(w, "No tickets found", http.StatusNotFound)
		return
	}

	// Kirimkan hasil dalam bentuk JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tiketList)
}

func GetTransaksiPenyelenggara(w http.ResponseWriter, r *http.Request) {
	// Ambil Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		return
	}

	// Ekstrak token dari "Bearer <token>"
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Klaim untuk menyimpan informasi dari token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil // Sesuaikan `jwtKey` dengan secret key JWT Anda
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Ambil user_id dari klaim
	userID := claims.UserID // Pastikan klaim `UserID` sesuai dengan struktur Anda
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	// Query untuk mendapatkan transaksi penyelenggara
	query := `
	  SELECT
		t.transaksi_id,
		t.tiket_id,
		u_penyelenggara.user_id AS penyelenggara_id,
		u_penyelenggara.name AS penyelenggara_name,
		k.nama_konser AS konser_name,
		u_pembeli.user_id AS pembeli_id,
		u_pembeli.name AS pembeli_name,
		t.status AS transaksi_status,
		t.qrcode AS qr_code,
		t.created_at AS transaksi_date
	  FROM transaksi t
	  JOIN tiket tk ON t.tiket_id = tk.tiket_id
	  JOIN konser k ON tk.konser_id = k.konser_id
	  JOIN users u_penyelenggara ON k.user_id = u_penyelenggara.user_id
	  JOIN users u_pembeli ON t.user_id = u_pembeli.user_id
	  WHERE u_penyelenggara.user_id = $1
	  ORDER BY t.created_at DESC;
	`

	// Jalankan query
	rows, err := config.DB.Query(query, userID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parsing hasil query
	var transaksiList []model.TransaksiPenyelenggara
	for rows.Next() {
		var transaksi model.TransaksiPenyelenggara
		if err := rows.Scan(
			&transaksi.TransaksiID,
			&transaksi.TiketID,
			&transaksi.PenyelenggaraID,
			&transaksi.PenyelenggaraName,
			&transaksi.KonserName,
			&transaksi.PembeliID,
			&transaksi.PembeliName,
			&transaksi.TransaksiStatus,
			&transaksi.QRCode,
			&transaksi.TransaksiDate,
		); err != nil {
			http.Error(w, "Failed to scan result: "+err.Error(), http.StatusInternalServerError)
			return
		}
		transaksiList = append(transaksiList, transaksi)
	}

	// Cek apakah ada data
	if len(transaksiList) == 0 {
		http.Error(w, "No transactions found", http.StatusNotFound)
		return
	}

	// Kirimkan hasil dalam bentuk JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaksiList)
}
