package main

import (
	"encoding/json" // Package baru untuk encoding JSON
	"fmt"
	"log"
	"net/http"

	_ "doc6/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// Struktur data untuk respons JSON
type ResponTest struct {
	Nilai int `json:"nilai"` // Tag 'json:"nilai"' memastikan outputnya adalah '{"nilai":...}'
}

// handlerHome godoc
// @Summary Home endpoint
// @Description Menampilkan pesan selamat datang
// @Tags home
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Internal Server Error"
// @Router / [get]
func handlerHome(w http.ResponseWriter, r *http.Request) {
	// Memeriksa metode
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	// Simulasi error 500 (contoh, bisa dihapus jika tidak ingin error random)
	// if rand.Intn(10) == 0 {
	//      http.Error(w, "Terjadi kesalahan server", http.StatusInternalServerError)
	//      return
	// }
	fmt.Fprintf(w, "Selamat datang di API GoLang Sederhana!\n")
	fmt.Fprintf(w, "Anda berhasil mengakses endpoint: %s\n", r.URL.Path)
}

// handlerApiTest godoc
// @Summary Test API
// @Description Mendapatkan nilai contoh
// @Tags test
// @Success 200 {object} ResponTest "Berhasil mendapatkan nilai"
// @Failure 500 {string} string "Gagal meng-encode JSON"
// @Router /api/test [get]
func handlerApiTest(w http.ResponseWriter, r *http.Request) {
	// Memeriksa metode
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// 1. Membuat data respons
	data := ResponTest{
		Nilai: 2,
	}

	// 2. Mengatur Header Content-Type
	w.Header().Set("Content-Type", "application/json")

	// 3. Encode data ke JSON dan menuliskannya ke http.ResponseWriter
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Gagal meng-encode JSON", http.StatusInternalServerError)
		log.Println("Error encoding JSON:", err)
		return
	}
	// Status HTTP 200 OK sudah otomatis diatur jika tidak ada error.
}

// @title Simple API
// @version 1.0
// @description API sederhana Golang
// @host localhost:8080
// @BasePath /
func main() {
	// 1. Definisikan Router (Multiplexer)
	mux := http.NewServeMux()

	// 2. Daftarkan Handler
	mux.HandleFunc("/", handlerHome)
	// Pendaftaran endpoint baru
	mux.HandleFunc("/api/test", handlerApiTest)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// 3. Konfigurasi dan Jalankan Server
	port := ":8080"
	log.Printf("Server siap berjalan di http://localhost%s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
