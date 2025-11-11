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

// handlerApiData (Mengambil 1 data terbaru - Handler Awal)
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
// --- Handler Baru: /api/showall ---
// ------------------------------------------

// handlerApiShowAll menangani endpoint "/api/showall" (Metode GET)
// Digunakan untuk mengambil SEMUA dokumen dari koleksi "alat"
func handlerApiShowAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	collection := mongoClient.Database(MongoDatabaseName).Collection(MongoCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Timeout lebih panjang untuk semua data
	defer cancel()

	// Membuat filter kosong untuk mendapatkan semua dokumen
	filter := bson.D{}

	// Mengambil kursor dari hasil pencarian
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Printf("Gagal mencari semua data dari MongoDB: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx) // Pastikan kursor ditutup

	var results []Alat // Slice untuk menampung semua hasil

	// Iterasi melalui hasil dan decode ke slice Alat
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Gagal mendekode semua dokumen: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}

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

// ------------------------------------------
// --- Handler Baru: /api/data/:rfid ---
// ------------------------------------------

// Fungsi utilitas untuk mengekstrak nilai RFID dari URL
// Misalnya dari "/api/data/a0822c23" akan menghasilkan "a0822c23"
func extractRFIDFromURL(path string) (string, error) {
	// Memisahkan path berdasarkan "/"
	parts := strings.Split(path, "/")

	// Kita mengharapkan format path: /api/data/rfid_value
	// Setelah split, kita akan punya: ["", "api", "data", "rfid_value"]
	if len(parts) < 4 || parts[3] == "" {
		return "", fmt.Errorf("format URL tidak valid atau RFID tidak ditemukan")
	}
	return parts[3], nil
}

// handlerApiDataByRFID menangani endpoint "/api/data/:rfid" (Metode GET)
// Digunakan untuk mengambil satu dokumen berdasarkan nilai field "rfid"
func handlerApiDataByRFID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// 1. Ekstrak RFID dari URL Path
	rfidValue, err := extractRFIDFromURL(r.URL.Path)
	if err != nil {
		http.Error(w, "RFID diperlukan: /api/data/{rfid_value}", http.StatusBadRequest)
		return
	}

	collection := mongoClient.Database(MongoDatabaseName).Collection(MongoCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Buat Filter berdasarkan RFID
	// Mencari dokumen di mana field "rfid" sama dengan rfidValue
	filter := bson.M{"rfid": rfidValue}

	var result Alat

	// 3. Mencari satu dokumen yang cocok dengan filter
	err = collection.FindOne(ctx, filter).Decode(&result)

	if err == mongo.ErrNoDocuments {
		http.Error(w, fmt.Sprintf("Data dengan RFID '%s' tidak ditemukan", rfidValue), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Gagal mengambil data dari MongoDB untuk RFID '%s': %v", rfidValue, err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}

	// 4. Mengatur Header dan mengirim respons JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Gagal meng-encode respons: %v", err)
		http.Error(w, "Kesalahan Server Internal", http.StatusInternalServerError)
		return
	}
}

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
	// 1. Inisiasi Koneksi MongoDB
	mongoClient, err = initMongoDB()
	if err != nil {
		log.Fatalf("❌ Fatal Error: Gagal koneksi ke MongoDB: %v", err)
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
	// Endpoint lama: 1 data terbaru
	mux.HandleFunc("/api/data", handlerApiData)

	// Endpoint baru: Semua data
	mux.HandleFunc("/api/showall", handlerApiShowAll)

	// Endpoint baru: Data berdasarkan RFID.
	// Menggunakan pattern matching karena http.ServeMux tidak mendukung path parameter.
	// Semua path yang diawali "/api/data/" akan masuk ke handler ini.
	mux.HandleFunc("/api/data/", handlerApiDataByRFID)

	// 4. Konfigurasi dan Jalankan Server
	port := ":8080"
	log.Printf("Server siap berjalan di http://localhost%s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
