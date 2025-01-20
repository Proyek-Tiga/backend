package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"project-tiket/config"
	"project-tiket/model"
	"project-tiket/service"

	"github.com/gorilla/mux"
	"github.com/veritrans/go-midtrans"
)

// CreatePayment untuk membuat pembayaran
func CreatePayment(w http.ResponseWriter, r *http.Request) {
	// Ambil transaksi_id dari parameter URL
	vars := mux.Vars(r)
	orderID := vars["transaksi_id"]

	// Inisialisasi struct paymentReq dan set transaksi_id
	var paymentReq model.Payment
	paymentReq.OrderID = orderID

	fmt.Println(paymentReq)

	// Decode request body untuk field lainnya
	if err := json.NewDecoder(r.Body).Decode(&paymentReq); err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	// Pastikan user_id tersedia
	if paymentReq.UserID == "" {
		http.Error(w, "User ID tidak disediakan", http.StatusBadRequest)
		return
	}

	// Ambil transaksi dari database berdasarkan OrderID yang diberikan
	var transaction model.Transaksi
	query := `
		SELECT transaksi_id, user_id, tiket_id, qty, harga, created_at, updated_at 
		FROM transaksi 
		WHERE transaksi_id = $1`
	err := config.DB.QueryRow(query, paymentReq.OrderID).Scan(
		&transaction.TransaksiID,
		&transaction.UserID,
		&transaction.TiketID,
		&transaction.Qty,
		&transaction.Harga,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Transaksi tidak ditemukan", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Error executing query: %v", err), http.StatusInternalServerError)
		return
	}

	// Hitung gross_amount
	paymentReq.GrossAmount = float64(transaction.Qty) * float64(transaction.Harga)
	fmt.Println(paymentReq.GrossAmount)

	// Inisialisasi Midtrans Client dan Snap Gateway
	midtransClient := service.MidtransClient()
	snapGateway := midtrans.SnapGateway{Client: *midtransClient}

	// Membuat request ke Midtrans dengan data transaksi
	snapReq := &midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  paymentReq.OrderID,
			GrossAmt: int64(paymentReq.GrossAmount * 1.5),
		},
	}

	// Dapatkan token dari Midtrans untuk pembayaran
	snapResp, err := snapGateway.GetToken(snapReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal membuat pembayaran: %v", err), http.StatusInternalServerError)
		return
	}

	// Tambahkan snap_url dan status pembayaran
	paymentReq.SnapURL = snapResp.RedirectURL
	paymentReq.Status = "Pending"
	paymentReq.CreatedAt = time.Now()

	// Simpan data pembayaran ke database
	insertQuery := `
		INSERT INTO payment (order_id, user_id, gross_amount, snap_url, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = config.DB.Exec(insertQuery, paymentReq.OrderID, paymentReq.UserID, paymentReq.GrossAmount, paymentReq.SnapURL, paymentReq.Status, paymentReq.CreatedAt)
	if err != nil {
		http.Error(w, "Gagal menyimpan pembayaran", http.StatusInternalServerError)
		return
	}

	// Kirim response dengan snap_url dan order_id
	response := map[string]interface{}{
		"snap_url": paymentReq.SnapURL,
		"order_id": paymentReq.OrderID,
		"status":   paymentReq.Status,
	}

	// Kirimkan response ke client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
