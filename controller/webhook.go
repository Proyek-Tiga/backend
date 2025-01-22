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
	if !ok {
	  http.Error(w, "Invalid order_id format", http.StatusBadRequest)
	  return
	}
  
  
	transactionStatus, ok := notification["transaction_status"].(string)
	if !ok {
	  http.Error(w, "Invalid status format", http.StatusBadRequest)
	  return
	}
  
  
	// Update the payment status in the database
	query := `UPDATE payment SET status = $1 WHERE order_id = $2`
	_, err := config.DB.Exec(query, transactionStatus, orderID)
	if err != nil {
	  http.Error(w, "Failed to update payment status: "+err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	// Respond to Midtrans with 200 OK to acknowledge receipt of the notification
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Notification processed successfully")
  }
    
