package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

// --- Handler Existing (Tanpa Perubahan) ---

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

// --- Handler Baru: /api/data ---

// handlerApiData menangani endpoint "/api/data" (Metode GET)
// Digunakan untuk mengambil satu dokumen terbaru dari koleksi "alat"
func handlerApiData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Mendapatkan koneksi ke koleksi MongoDB
	collection := mongoClient.Database(MongoDatabaseName).Collection(MongoCollectionName)

	// Membuat konteks dengan timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result Alat

	// Opsi untuk mencari dokumen terbaru (berdasarkan _id)
	// Sort descending by _id dan limit 1
	findOptions := options.FindOne().SetSort(bson.D{{"_id", -1}})

	// Mencari satu dokumen (yang terbaru)
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

	// Mengatur Header dan mengirim respons JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Gagal meng-encode respons: %v", err)
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

// --- Fungsi Main ---

func main() {
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

	// 2. Definisikan Router
	mux := http.NewServeMux()

	// 3. Daftarkan Handler
	mux.HandleFunc("/", handlerHome)
	mux.HandleFunc("/api/test", handlerApiTest)
	// Pendaftaran endpoint MongoDB baru
	mux.HandleFunc("/api/data", handlerApiData)

	// 4. Konfigurasi dan Jalankan Server
	port := ":8080"
	log.Printf("Server siap berjalan di http://localhost%s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
