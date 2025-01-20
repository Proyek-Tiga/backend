package service

import (
	"os"

	"github.com/veritrans/go-midtrans"
)

// MidtransClient menginisialisasi client Midtrans
func MidtransClient() *midtrans.Client {
	c := midtrans.NewClient()
	c.ServerKey = os.Getenv("MIDTRANS_SERVER_KEY")
	c.ClientKey = os.Getenv("MIDTRANS_CLIENT_KEY")
	c.APIEnvType = midtrans.Sandbox // Gunakan Sandbox untuk testing, ubah ke Production untuk live
	return &c                       // Mengembalikan pointer ke client
}
