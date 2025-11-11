package main

import (
	"fmt"
	"log"
	"net/http"
)

// handlerHome adalah fungsi yang akan menangani permintaan ke endpoint "/"
func handlerHome(w http.ResponseWriter, r *http.Request) {
	// Pastikan metode permintaan adalah GET
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Mengatur header konten sebagai JSON (meskipun kita hanya mengirim teks)
	// Untuk API nyata, Anda akan mengirim JSON di sini.
	w.Header().Set("Content-Type", "text/plain")

	// Mengatur status kode HTTP 200 OK secara default (jika tidak ada error)

	// Menulis respons ke client
	fmt.Fprintf(w, "Selamat datang di API GoLang Sederhana!\n")
	fmt.Fprintf(w, "Anda berhasil mengakses endpoint: %s\n", r.URL.Path)
}

func main() {
	// 1. Definisikan Router (Multiplexer)
	// http.NewServeMux() adalah router dasar di Go
	mux := http.NewServeMux()

	// 2. Daftarkan Handler untuk Endpoint "/"
	// Ketika ada permintaan ke "/", fungsi handlerHome akan dipanggil.
	mux.HandleFunc("/", handlerHome)

	// 3. Konfigurasi Server
	port := ":8080"
	log.Printf("Server siap berjalan di http://localhost%s", port)

	// 4. Jalankan Server
	// http.ListenAndServe(port, handler) akan memulai server HTTP.
	// Jika ada error (misalnya port sudah terpakai), ia akan mengembalikan error.
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
