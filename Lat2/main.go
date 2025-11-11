package main

import (
	"encoding/json" // Package baru untuk encoding JSON
	"fmt"
	"log"
	"net/http"
)

// Struktur data untuk respons JSON
type ResponTest struct {
	Nilai int `json:"nilai"` // Tag 'json:"nilai"' memastikan outputnya adalah '{"nilai":...}'
}

// handlerHome menangani endpoint "/" (Metode GET)
func handlerHome(w http.ResponseWriter, r *http.Request) {
	// Memeriksa metode
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Selamat datang di API GoLang Sederhana!\n")
	fmt.Fprintf(w, "Anda berhasil mengakses endpoint: %s\n", r.URL.Path)
}

// handlerApiTest menangani endpoint "/api/test" (Metode GET)
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
	// Ini memberitahu klien bahwa responsnya adalah JSON
	w.Header().Set("Content-Type", "application/json")

	// 3. Encode data ke JSON dan menuliskannya ke http.ResponseWriter
	// json.NewEncoder(w) akan menuliskan output JSON langsung ke respons HTTP
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Gagal meng-encode JSON", http.StatusInternalServerError)
		log.Println("Error encoding JSON:", err)
		return
	}
	// Status HTTP 200 OK sudah otomatis diatur jika tidak ada error.
}

func main() {
	// 1. Definisikan Router (Multiplexer)
	mux := http.NewServeMux()

	// 2. Daftarkan Handler
	mux.HandleFunc("/", handlerHome)
	// Pendaftaran endpoint baru
	mux.HandleFunc("/api/test", handlerApiTest)

	// 3. Konfigurasi dan Jalankan Server
	port := ":8080"
	log.Printf("Server siap berjalan di http://localhost%s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
