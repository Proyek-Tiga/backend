package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"project-tiket/config"
	"project-tiket/model"

	"github.com/gorilla/mux"
	// "github.com/gorilla/mux"
)

func AddKonser(w http.ResponseWriter, r *http.Request) {

	// Parse form-data
	err := r.ParseMultipartForm(10 << 20) // Maksimum ukuran file 10MB
	if err != nil {
		http.Error(w, "Unable to parse form-data: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ambil data teks dari form-data
	userID := r.FormValue("user_id")
	lokasiID := r.FormValue("lokasi_id")
	namaKonser := r.FormValue("nama_konser")
	tanggalKonser := r.FormValue("tanggal_konser")
	harga, _ := strconv.Atoi(r.FormValue("harga"))

	// Ambil file dari form-data
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Unable to retrieve the image: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Baca data file gambar
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read the image file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 1: Upload gambar ke GitHub
	githubToken := os.Getenv("GH_ACCESS_TOKEN") // Ganti dengan token Anda
	repoOwner := "Proyek-Tiga"                  // Nama organisasi GitHub
	repoName := "images"                        // Nama repositori
	filePath := fmt.Sprintf("konser/%d_%s.jpg", time.Now().Unix(), namaKonser)
	uploadURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", repoOwner, repoName, filePath)

	encodedImage := base64.StdEncoding.EncodeToString(fileBytes)
	payload := map[string]string{
		"message": fmt.Sprintf("Add image for concert %s", namaKonser),
		"content": encodedImage,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PUT", uploadURL, bytes.NewReader(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to upload image to GitHub: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("GitHub API error: %s", string(body)), http.StatusInternalServerError)
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	imageURL := result["content"].(map[string]interface{})["download_url"].(string)

	// Step 2: Ambil jumlah tiket dari lokasi
	var jumlahTiket int
	err = config.DB.QueryRow("SELECT tiket FROM lokasi WHERE lokasi_id = $1", lokasiID).Scan(&jumlahTiket)
	if err != nil {
		http.Error(w, "Failed to retrieve ticket quantity from location: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 3: Simpan URL gambar dan data konser ke database
	query := `
	INSERT INTO konser (user_id, lokasi_id, nama_konser, tanggal_konser, jumlah_tiket, harga, image, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', NOW(), NOW())
	RETURNING konser_id`

	var id string
	err = config.DB.QueryRow(query,
		userID,
		lokasiID,
		namaKonser,
		tanggalKonser,
		jumlahTiket, // Menggunakan jumlah tiket dari lokasi
		harga,
		imageURL, // URL gambar dari GitHub
	).Scan(&id)

	if err != nil {
		http.Error(w, "Failed to save concert data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the newly created ID and image URL in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Konser added successfully",
		"id":      id,
		"image":   imageURL,
	})
}

func UpdateKonser(w http.ResponseWriter, r *http.Request) {
	// Parse form-data
	err := r.ParseMultipartForm(10 << 20) // Maksimum ukuran file 10MB
	if err != nil {
		http.Error(w, "Unable to parse form-data: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse ID from the URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Ambil data teks dari form-data
	userID := r.FormValue("user_id")
	lokasiID := r.FormValue("lokasi_id")
	namaKonser := r.FormValue("nama_konser")
	tanggalKonser := r.FormValue("tanggal_konser")
	harga, _ := strconv.Atoi(r.FormValue("harga"))

	// Ambil file dari form-data (gambar)
	file, fileHeader, err := r.FormFile("image")
	var imageURL, status string

	// Ambil status konser dari database
	err = config.DB.QueryRow("SELECT image, status FROM konser WHERE konser_id = $1", id).Scan(&imageURL, &status)
	if err != nil {
		http.Error(w, "Failed to retrieve existing data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Jika gambar baru diunggah
	if err != http.ErrMissingFile && fileHeader != nil {
		defer file.Close()

		// Baca data file gambar
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, "Unable to read the image file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Step 1: Upload gambar ke GitHub
		githubToken := os.Getenv("GH_ACCESS_TOKEN") // Ganti dengan token Anda
		repoOwner := "Proyek-Tiga"                  // Nama organisasi GitHub
		repoName := "images"                        // Nama repositori
		filePath := fmt.Sprintf("konser/%d_%s.jpg", time.Now().Unix(), namaKonser)
		uploadURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", repoOwner, repoName, filePath)

		encodedImage := base64.StdEncoding.EncodeToString(fileBytes)
		payload := map[string]string{
			"message": fmt.Sprintf("Update image for concert %s", namaKonser),
			"content": encodedImage,
		}
		payloadBytes, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PUT", uploadURL, bytes.NewReader(payloadBytes))
		req.Header.Set("Authorization", "Bearer "+githubToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to upload image to GitHub: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := ioutil.ReadAll(resp.Body)
			http.Error(w, fmt.Sprintf("GitHub API error: %s", string(body)), http.StatusInternalServerError)
			return
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		imageURL = result["content"].(map[string]interface{})["download_url"].(string)
	}

	// Step 2: Update jumlah tiket berdasarkan lokasi
	var jumlahTiket int
	err = config.DB.QueryRow("SELECT tiket FROM lokasi WHERE lokasi_id = $1", lokasiID).Scan(&jumlahTiket)
	if err != nil {
		http.Error(w, "Failed to retrieve ticket quantity from location: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 3: Update data konser di database
	query := `
	  UPDATE konser
	  SET user_id = $1,
		lokasi_id = $2,
		nama_konser = $3,
		tanggal_konser = $4,
		jumlah_tiket = $5,
		harga = $6,
		image = $7,
		status = $8,
		updated_at = NOW()
	  WHERE konser_id = $9`

	_, err = config.DB.Exec(query,
		userID,
		lokasiID,
		namaKonser,
		tanggalKonser,
		jumlahTiket, // Menggunakan jumlah tiket dari lokasi
		harga,
		imageURL, // URL gambar dari GitHub atau gambar lama dari database
		status,   // Menggunakan status lama dari database
		id,       // ID konser yang akan diupdate
	)

	if err != nil {
		http.Error(w, "Failed to update concert data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Konser updated successfully",
		"id":      id,
		"image":   imageURL,
	})
}

func GetAllKonser(w http.ResponseWriter, r *http.Request) {
	// Query untuk mengambil data konser, user, dan lokasi
	query := `
		SELECT k.user_id, k.lokasi_id, k.konser_id, k.nama_konser, k.tanggal_konser, k.jumlah_tiket, k.harga, k.image, k.status, 
		       k.created_at, k.updated_at, 
		       u.name, l.lokasi
		FROM konser k
		INNER JOIN users u ON k.user_id = u.user_id
		INNER JOIN lokasi l ON k.lokasi_id = l.lokasi_id`

	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var konserList []model.Konser
	for rows.Next() {
		var konser model.Konser
		var userName, lokasiName string

		// Scan data konser, user_name, lokasi_name
		if err := rows.Scan(
			&konser.UserID,
			&konser.LokasiID,
			&konser.KonserID,
			&konser.NamaKonser,
			&konser.TanggalKonser,
			&konser.JumlahTiket,
			&konser.Harga,
			&konser.Image,
			&konser.Status,
			&konser.CreatedAt,
			&konser.UpdatedAt,
			&userName,
			&lokasiName,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Menambahkan nama user dan lokasi ke struct konser
		konser.UserName = userName
		konser.LokasiName = lokasiName

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
		`SELECT k.user_id, k.lokasi_id, k.konser_id, k.nama_konser, k.tanggal_konser, k.jumlah_tiket, k.harga, k.image, k.status, 
		       k.created_at, k.updated_at, 
		       u.name, l.lokasi
		FROM konser k
		INNER JOIN users u ON k.user_id = u.user_id
		INNER JOIN lokasi l ON k.lokasi_id = l.lokasi_id WHERE k.konser_id = $1`,
		id,
	).Scan(
		&konser.UserID,
		&konser.LokasiID,
		&konser.KonserID,
		&konser.NamaKonser,
		&konser.TanggalKonser,
		&konser.JumlahTiket,
		&konser.Harga,
		&konser.Image,
		&konser.Status,
		&konser.CreatedAt,
		&konser.UpdatedAt,
		&konser.UserName,
		&konser.LokasiName,
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

	// // Jika status yang diberikan adalah "approved", lakukan penanganan lebih lanjut (misalnya menambah tiket atau tindakan lainnya)
	// if status.Status == "approved" {
	// 	// Anda bisa menambahkan logika lebih lanjut untuk "approved" di sini
	// 	// Misalnya, menambahkan data ke tabel lain jika perlu
	// }

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
	var concerts []struct {
		model.Konser
		Lokasi model.Lokasi `json:"lokasi"`
	}

	// Query untuk mengambil data konser dengan status "approved" beserta data lokasi
	query := `
		SELECT 
			k.konser_id, k.user_id, k.lokasi_id, k.nama_konser, k.tanggal_konser, k.jumlah_tiket, 
			k.harga, k.image, k.status, 
			k.created_at, k.updated_at,
			l.lokasi_id, l.lokasi, l.tiket, l.created_at, l.updated_at
		FROM konser k
		JOIN lokasi l ON k.lokasi_id = l.lokasi_id
		WHERE k.status = $1`

	rows, err := config.DB.Query(query, "approved")
	if err != nil {
		http.Error(w, "Failed to fetch approved concerts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var concert struct {
			model.Konser
			Lokasi model.Lokasi `json:"lokasi"`
		}

		err := rows.Scan(
			&concert.Konser.KonserID,
			&concert.Konser.UserID,
			&concert.Konser.LokasiID,
			&concert.Konser.NamaKonser,
			&concert.Konser.TanggalKonser,
			&concert.Konser.JumlahTiket,
			&concert.Konser.Harga,
			&concert.Konser.Image,
			&concert.Konser.Status,
			&concert.Konser.CreatedAt,
			&concert.Konser.UpdatedAt,
			&concert.Lokasi.LokasiID,
			&concert.Lokasi.Lokasi,
			&concert.Lokasi.Tiket,
			&concert.Lokasi.CreatedAt,
			&concert.Lokasi.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Failed to scan concert data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		concerts = append(concerts, concert)
	}

	if len(concerts) == 0 {
		http.Error(w, "No approved concerts found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(concerts)
}

func DeleteKonser(w http.ResponseWriter, r *http.Request) {
	// Ambil konser_id dari URL parameter
	vars := mux.Vars(r)
	konserID, ok := vars["id"]
	if !ok {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}

	// Mulai transaksi database
	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, "Gagal memulai transaksi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Hapus tiket yang terkait dengan konser
	deleteTiketQuery := `
	  DELETE FROM tiket
	  WHERE konser_id = $1`
	_, err = tx.Exec(deleteTiketQuery, konserID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Gagal menghapus tiket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Hapus konser
	deleteKonserQuery := `
	  DELETE FROM konser
	  WHERE konser_id = $1`
	_, err = tx.Exec(deleteKonserQuery, konserID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Gagal menghapus konser: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit transaksi
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Gagal menyelesaikan transaksi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kirim response ke client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Konser dengan ID %s dan tiket terkait berhasil dihapus", konserID),
	})
}
