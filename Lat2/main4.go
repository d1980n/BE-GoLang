package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- Konfigurasi MongoDB ---
const (
	MongoConnectionString = "mongodb://kawal_anak:1hoUMt847hO4pgi@nosql.smartsystem.id:27017/kawal_anak"
	MongoDatabaseName     = "kawal_anak"
	MongoCollectionName   = "alat"
)

const (
	// ... (konfigurasi MongoDB)

	// Konfigurasi FTP (Ganti dengan detail FTP Anda)
	FtpServerHost = "172.31.185.190" // Ganti dengan host dan port FTP Anda
	FtpUsername   = "dan"            // Ganti dengan username FTP Anda
	FtpPassword   = "123"            // Ganti dengan password FTP Anda
	FtpRemoteDir  = "/upload"        // Direktori tujuan di server FTP

	LocalUploadDir = "./Picture" // Direktori lokal untuk fallback
)

// Variabel global untuk klien MongoDB
var mongoClient *mongo.Client

// --- Struktur Data ---

// Struktur untuk endpoint /api/test
type ResponTest struct {
	Nilai int `json:"nilai"`
}

// Struktur untuk dokumen koleksi "alat" di MongoDB
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

// ------------------------------------------
// --- Middleware CORS ---
// ------------------------------------------
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

// --- Handler Existing ---

// handlerHome menangani endpoint "/" (Metode GET)
func handlerHome(w http.ResponseWriter, r *http.Request) {
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
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}
	data := ResponTest{Nilai: 2}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Gagal meng-encode JSON", http.StatusInternalServerError)
		log.Println("Error encoding JSON:", err)
		return
	}
}

// handlerApiData (Mengambil 1 data terbaru)
func handlerApiData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	collection := mongoClient.Database(MongoDatabaseName).Collection(MongoCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result Alat

	findOptions := options.FindOne().SetSort(bson.D{{"_id", -1}})

	err := collection.FindOne(ctx, bson.D{}, findOptions).Decode(&result)

	if err == mongo.ErrNoDocuments {
		http.Error(w, "Data tidak ditemukan di koleksi 'alat'", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Gagal mengambil data dari MongoDB: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Gagal meng-encode respons: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}
}

// ------------------------------------------
// --- Handler: /api/showall & /api/dataall ---
// ------------------------------------------

// handlerApiShowAll menangani endpoint "/api/showall" dan "/api/dataall" (Metode GET)
func handlerApiShowAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	collection := mongoClient.Database(MongoDatabaseName).Collection(MongoCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	filter := bson.D{}

	// Opsi untuk mengurutkan berdasarkan ID turun (terbaru dulu)
	findOptions := options.Find().SetSort(bson.D{{"_id", -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("Gagal mencari semua data dari MongoDB: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var results []Alat

	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Gagal mendekode semua dokumen: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(results) == 0 {
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

// ------------------------------------------
// --- Handler: /api/data/:rfid ---
// ------------------------------------------

func extractRFIDFromURL(path string) (string, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[3] == "" {
		return "", fmt.Errorf("format URL tidak valid atau RFID tidak ditemukan")
	}
	return parts[3], nil
}

// handlerApiDataByRFID menangani endpoint "/api/data/:rfid" (Metode GET)
func handlerApiDataByRFID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	rfidValue, err := extractRFIDFromURL(r.URL.Path)
	if err != nil {
		http.Error(w, "RFID diperlukan: /api/data/{rfid_value}", http.StatusBadRequest)
		return
	}

	collection := mongoClient.Database(MongoDatabaseName).Collection(MongoCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Hanya ambil data terbaru untuk RFID ini
	filter := bson.M{"rfid": rfidValue}
	findOptions := options.FindOne().SetSort(bson.D{{"_id", -1}}) // Tambahkan pengurutan

	var result Alat

	err = collection.FindOne(ctx, filter, findOptions).Decode(&result)

	if err == mongo.ErrNoDocuments {
		http.Error(w, fmt.Sprintf("Data dengan RFID '%s' tidak ditemukan", rfidValue), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Gagal mengambil data dari MongoDB untuk RFID '%s': %v", rfidValue, err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Gagal meng-encode respons: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}
}

// ------------------------------------------
// --- Handler: /api/upload
// ------------------------------------------

// --- Fungsi Koneksi MongoDB ---

func initMongoDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoConnectionString))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat klien MongoDB: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("gagal melakukan ping ke MongoDB: %w", err)
	}

	log.Println("✅ Berhasil terhubung ke MongoDB!")
	return client, nil
}

// --- Fungsi Main ---

func main() {
	var err error

	// 1. Inisialisasi MongoDB
	mongoClient, err = initMongoDB()
	if err != nil {
		log.Fatalf("❌ Fatal Error: Gagal koneksi ke MongoDB: %v", err)
	}
	defer func() {
		if err = mongoClient.Disconnect(context.TODO()); err != nil {
			log.Printf("Error saat memutuskan koneksi MongoDB: %v", err)
		}
	}()

	// 2. Definisikan Router
	mux := http.NewServeMux()

	// 3. Daftarkan Handler dengan membungkusnya menggunakan middleware enableCORS

	// Endpoint "/"
	mux.HandleFunc("/", enableCORS(handlerHome))

	// Endpoint "/api/test"
	mux.HandleFunc("/api/test", enableCORS(handlerApiTest))

	// Endpoint "/api/data" (terbaru)
	mux.HandleFunc("/api/data", enableCORS(handlerApiData))

	// Endpoint "/api/showall" (semua data - existing)
	mux.HandleFunc("/api/showall", enableCORS(handlerApiShowAll))

	// Endpoint BARU "/api/dataall" (semua data - menggunakan handler yang sama)
	mux.HandleFunc("/api/dataall", enableCORS(handlerApiShowAll))

	// Endpoint "/api/data/:rfid"
	mux.HandleFunc("/api/data/", enableCORS(handlerApiDataByRFID))

	// ----------------------------------------------------
	// --- Static File Server untuk Gambar (BARU) ---
	// ----------------------------------------------------
	// Handler untuk menyajikan file statis dari folder "picture"
	// Akses gambar melalui path "/pictures/{nama_file.jpg}"
	// Contoh: http://0.0.0.0:8080/pictures/gambar_rfid1.jpg

	// 1. Definisikan File Server yang akan mencari file di folder "picture"
	fileServer := http.FileServer(http.Dir("./picture"))

	// 2. Bungkus logic StripPrefix ke dalam HandlerFunc dan terapkan middleware CORS
	mux.HandleFunc("/pictures/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		// http.StripPrefix menghapus bagian "/pictures/" dari path URL sebelum mencari file di disk.
		// Contoh: Permintaan /pictures/foto.jpg akan dicari sebagai ./picture/foto.jpg
		http.StripPrefix("/pictures/", fileServer).ServeHTTP(w, r)
	}))
	log.Println("✅ Static File Server aktif di /pictures/")

	// 4. Konfigurasi dan Jalankan Server
	// Menggunakan "0.0.0.0:8080" secara eksplisit untuk menghindari masalah binding IP
	port := "0.0.0.0:8080"
	log.Printf("Server siap berjalan di http://%s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
