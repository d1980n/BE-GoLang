package main

import (
	// "errors" // Kita tidak perlu ini lagi untuk metode baru
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings" // Pastikan "strings" di-import
	"time"

	"github.com/jlaffaye/ftp"
)

// --- Konfigurasi FTP ---
const (
	ftpHost     = "172.31.185.190"
	ftpPort     = "21"
	ftpUser     = "dan"
	ftpPass     = "123"
	ftpBasePath = "/upload"
)

func pictureHandler(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/api/picture/")
	if filename == "" {
		http.Error(w, "Nama file tidak boleh kosong", http.StatusBadRequest)
		return
	}

	log.Printf("Mencoba mengambil file: %s", filename)

	c, err := ftp.Dial(ftpHost+":"+ftpPort, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Printf("Gagal koneksi ke FTP server: %v", err)
		http.Error(w, "Kesalahan internal server (FTP Dial)", http.StatusInternalServerError)
		return
	}
	defer c.Quit()

	if err = c.Login(ftpUser, ftpPass); err != nil {
		log.Printf("Gagal login FTP: %v", err)
		http.Error(w, "Kesalahan internal server (FTP Login)", http.StatusInternalServerError)
		return
	}

	if err = c.ChangeDir(ftpBasePath); err != nil {
		log.Printf("Gagal pindah direktori FTP: %v", err)
		http.Error(w, "Kesalahan internal server (FTP CWD)", http.StatusInternalServerError)
		return
	}

	// Ini adalah bagian "if err != nil" yang kita maksud
	resp, err := c.Retr(filename)
	if err != nil {

		// ----- INI BAGIAN YANG DIPERBAIKI (METODE BARU) -----

		// Kita akan cek isi errornya, ini lebih aman
		// Kode FTP 550 berarti "File unavailable"
		if strings.Contains(err.Error(), "550") {
			log.Printf("File tidak ditemukan di FTP (Error 550): %s", filename)
			http.Error(w, "File tidak ditemukan", http.StatusNotFound)
		} else {
			// Error lain
			log.Printf("Gagal mengambil file dari FTP (RETR): %v", err)
			http.Error(w, "Kesalahan internal server (FTP RETR)", http.StatusInternalServerError)
		}
		// ----- AKHIR BAGIAN PERBAIKAN -----

		return
	}
	defer resp.Close()

	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+filename+"\"")

	_, err = io.Copy(w, resp)
	if err != nil {
		log.Printf("Gagal mengirim file ke klien: %v", err)
	}

	log.Printf("Sukses mengirim file: %s", filename)
}

// --- Akhir Konfigurasi ---

func main() {
	http.HandleFunc("/api/picture/", pictureHandler)
	fmt.Println("🚀 Server dimulai di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
