package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"strings"

	"github.com/jlaffaye/ftp"
)

// --- Konfigurasi MongoDB ---
const (
	MongoConnectionString = "mongodb://kawal_anak:1hoUMt847hO4pgi@nosql.smartsystem.id:27017/kawal_anak"
	MongoDatabaseName     = "kawal_anak"
	MongoCollectionName   = "alat"
)

// --- Konfigurasi FTP ---
const (
	// -- ftpHost     = "10.222.254.191" ---
	ftpHost     = "10.200.9.191"
	ftpPort     = "21"
	ftpUser     = "dan"
	ftpPass     = "123"
	ftpBasePath = "/upload"
)

// Variabel global untuk klien MongoDB
var mongoClient *mongo.Client

type Alat struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	RFID               string             `bson:"rfid" json:"rfid"`
	Weight             float64            `bson:"weight" json:"weight"`
	Height             float64            `bson:"height" json:"height"`
	Pict1URL           string             `bson:"pict1_url" json:"pict1_url"`
	Pict2URL           string             `bson:"pict2_url" json:"pict2_url"`
	Pict3URL           string             `bson:"pict3_url" json:"pict3_url"`
	IngestionTimestamp time.Time          `bson:"ingestion_timestamp" json:"ingestion_timestamp"`
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mengizinkan Origin mana pun untuk mengakses resource
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Mengizinkan metode-metode HTTP umum
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

		// Mengizinkan header yang mungkin digunakan oleh klien (Content-Type)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Jika permintaan adalah OPTIONS (pre-flight request), kirim respons 200 OK dan hentikan eksekusi
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Lanjutkan ke handler berikutnya
		next.ServeHTTP(w, r)
	}
}

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

func handlerApiData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Mendapatkan koneksi ke koleksi MongoDB
	collection := mongoClient.Database(MongoDatabaseName).Collection(MongoCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.D{}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Printf("Gagal mencari semua data dari MongoDB: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx) // Pastikan kursor ditutup

	var results []Alat

	// Mencari satu dokumen (yang terbaru)
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Gagal mendekode semua dokumen: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}

	// === PERUBAHAN DIMULAI DI SINI ===
	// Tentukan base URL untuk gambar
	// Tambahkan "http://" agar menjadi URL yang valid
	const imageBaseURL = "http://10.222.254.9:8080/api/picture/"

	// Iterasi melalui hasil dan ubah field URL gambar
	// Ini hanya mengubah data *sebelum* dikirim sebagai JSON,
	// tidak mengubah data di database.
	for i := range results {
		// Hanya tambahkan prefix jika nama file tidak kosong
		if results[i].Pict1URL != "" {
			results[i].Pict1URL = imageBaseURL + results[i].Pict1URL
		}
		if results[i].Pict2URL != "" {
			results[i].Pict2URL = imageBaseURL + results[i].Pict2URL
		}
		if results[i].Pict3URL != "" {
			results[i].Pict3URL = imageBaseURL + results[i].Pict3URL
		}
	}
	// === PERUBAHAN SELESAI DI SINI ===

	// Mengatur Header dan mengirim respons JSON
	w.Header().Set("Content-Type", "application/json")
	if len(results) == 0 {
		// Mengirim array kosong jika tidak ada dokumen
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("Gagal meng-encode respons showall: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}
}

// --- Fungsi Koneksi MongoDB ---

func initMongoDB() (*mongo.Client, error) {
	// Konteks dengan timeout untuk koneksi
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Membuat klien MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoConnectionString))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat klien MongoDB: %w", err)
	}

	// Menguji koneksi
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("gagal melakukan ping ke MongoDB: %w", err)
	}

	log.Println("Berhasil terhubung ke MongoDB!")
	return client, nil
}

func main() {
	// 1. Inisialisasi Klien MongoDB
	var err error
	// 1. Inisiasi Koneksi MongoDB
	mongoClient, err = initMongoDB()
	if err != nil {
		log.Fatalf("Fatal Error: Gagal koneksi ke MongoDB: %v", err)
	}
	// Pastikan koneksi ditutup saat aplikasi berhenti
	defer func() {
		if err = mongoClient.Disconnect(context.TODO()); err != nil {
			log.Printf("Error saat memutuskan koneksi MongoDB: %v", err)
		}
	}()

	mux := http.NewServeMux()

	//mux.HandleFunc("/", handlerHome)
	mux.HandleFunc("/api/datalast", enableCORS(handlerApiData))
	mux.HandleFunc("/api/picture/", enableCORS(pictureHandler))
	// 3. Konfigurasi Server
	port := ":8080"
	log.Printf("Server siap berjalan di http://localhost%s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
