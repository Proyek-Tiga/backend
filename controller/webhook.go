package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"project-tiket/config"

	_ "github.com/lib/pq" // Driver PostgreSQL
)

func MidtransWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Decode the incoming request body
	var notification map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "Invalid notification payload", http.StatusBadRequest)
		return
	}

	// Extract relevant fields from the notification
	orderID, ok := notification["order_id"].(string)
	if !ok || orderID == "" {
		http.Error(w, "Invalid or missing order_id", http.StatusBadRequest)
		return
	}

	transactionStatus, ok := notification["transaction_status"].(string)
	if !ok || transactionStatus == "" {
		http.Error(w, "Invalid or missing transaction_status", http.StatusBadRequest)
		return
	}

	// Update the payment status in the database (payment table)
	queryPayment := `UPDATE payment SET status = $1 WHERE order_id = $2`
	_, err := config.DB.Exec(queryPayment, transactionStatus, orderID)
	if err != nil {
		http.Error(w, "Failed to update payment status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the transaction status in the database (transaksi table)
	queryTransaksi := `UPDATE transaksi SET status = $1 WHERE transaksi_id = $2`
	_, err = config.DB.Exec(queryTransaksi, transactionStatus, orderID)
	if err != nil {
		http.Error(w, "Failed to update transaction status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond to Midtrans with 200 OK to acknowledge receipt of the notification
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Notification processed successfully")
}
