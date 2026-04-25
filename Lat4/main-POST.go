package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jlaffaye/ftp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Struktur request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Struktur response
type LoginResponse struct {
	Message string `json:"message"`
}

// Struktur request untuk edit register (dengan guid)
type EditRegRequest struct {
	Guid     string `json:"guid"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Struktur data user untuk MongoDB
type User struct {
	Guid     string `json:"guid" bson:"guid"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

// Struktur data picture untuk MongoDB
type Picture struct {
	Guid     string `json:"guid" bson:"guid"`
	Filename string `json:"filename" bson:"filename"`
	Timenow  string `json:"timenow" bson:"timenow"`
}

// MongoDB config
const (
	MongoConnectionString = "mongodb://kawal_anak:1hoUMt847hO4pgi@nosql.smartsystem.id:27017/kawal_anak"
	MongoDatabaseName     = "kawal_anak"
	MongoCollectionName   = "users"
)

// FTP config
const (
	ftpHost     = "172.31.185.190"
	ftpPort     = "21"
	ftpUser     = "dan"
	ftpPass     = "123"
	ftpBasePath = "/upload"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// hanya izinkan POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Method not allowed"})
		return
	}

	var req LoginRequest

	// decode JSON dari body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Invalid JSON"})
		return
	}

	// validasi input
	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Username dan password harus diisi"})
		return
	}

	// hash password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal hash password"})
		return
	}

	// koneksi ke MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(MongoConnectionString)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal koneksi ke database: " + err.Error()})
		return
	}
	defer client.Disconnect(ctx)

	// verifikasi koneksi
	err = client.Ping(ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal ping database: " + err.Error()})
		return
	}

	// simpan user ke MongoDB
	collection := client.Database(MongoDatabaseName).Collection(MongoCollectionName)
	user := User{
		Guid:     uuid.New().String(),
		Username: req.Username,
		Password: string(hashedPassword),
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal menyimpan data user: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(LoginResponse{Message: "Register berhasil"})
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	// hanya izinkan POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Method not allowed"})
		return
	}

	var req LoginRequest

	// decode JSON dari body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Invalid JSON"})
		return
	}

	// validasi sederhana
	if req.Username == "admin" && req.Password == "1234" {
		json.NewEncoder(w).Encode(LoginResponse{Message: "Login berhasil"})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Username/password salah"})
	}
}

func addpictHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// hanya izinkan POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Method not allowed"})
		return
	}

	// parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal parse form: " + err.Error()})
		return
	}

	// ambil file dari form
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "File tidak ditemukan: " + err.Error()})
		return
	}
	defer file.Close()

	// generate guid dan timenow
	pictGuid := uuid.New().String()
	timenow := time.Now().Format("2006-01-02 15:04:05")
	filename := handler.Filename

	// upload file ke FTP server
	ftpAddr := ftpHost + ":" + ftpPort
	conn, err := ftp.Dial(ftpAddr, ftp.DialWithTimeout(10*time.Second))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal koneksi ke FTP: " + err.Error()})
		return
	}
	defer conn.Quit()

	err = conn.Login(ftpUser, ftpPass)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal login FTP: " + err.Error()})
		return
	}

	// upload file ke path: /upload/filename
	remotePath := filepath.ToSlash(filepath.Join(ftpBasePath, filename))
	err = conn.Stor(remotePath, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal upload file ke FTP: " + err.Error()})
		return
	}

	// simpan metadata ke MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(MongoConnectionString)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal koneksi ke database: " + err.Error()})
		return
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal ping database: " + err.Error()})
		return
	}

	collection := client.Database(MongoDatabaseName).Collection(MongoCollectionName)
	pict := Picture{
		Guid:     pictGuid,
		Filename: filename,
		Timenow:  timenow,
	}

	_, err = collection.InsertOne(ctx, pict)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal menyimpan data picture: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(LoginResponse{Message: "Upload berhasil, file: " + filename})
}

// editRegHandler - POST /api1/edit-reg
// Edit data user (username & password) berdasarkan GUID
func editRegHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// hanya izinkan POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Method not allowed"})
		return
	}

	var req EditRegRequest

	// decode JSON dari body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Invalid JSON: " + err.Error()})
		return
	}

	// validasi: guid, username, dan password harus diisi
	if req.Guid == "" || req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Guid, username, dan password harus diisi"})
		return
	}

	// hash password baru
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal hash password"})
		return
	}

	// koneksi ke MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(MongoConnectionString)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal koneksi ke database: " + err.Error()})
		return
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal ping database: " + err.Error()})
		return
	}

	// update user berdasarkan GUID
	collection := client.Database(MongoDatabaseName).Collection(MongoCollectionName)

	filter := bson.M{"guid": req.Guid}
	update := bson.M{
		"$set": bson.M{
			"username": req.Username,
			"password": string(hashedPassword),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal update user: " + err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(LoginResponse{Message: "User dengan GUID tersebut tidak ditemukan"})
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{Message: "Edit register berhasil"})
}

// editRegPutHandler - PUT /api2/edit-reg
// Edit data user (username & password) berdasarkan GUID menggunakan PUT
func editRegPutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// hanya izinkan PUT
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Method not allowed, gunakan PUT"})
		return
	}

	var req EditRegRequest

	// decode JSON dari body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Invalid JSON: " + err.Error()})
		return
	}

	// validasi: guid, username, dan password harus diisi
	if req.Guid == "" || req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Guid, username, dan password harus diisi"})
		return
	}

	// hash password baru
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal hash password"})
		return
	}

	// koneksi ke MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(MongoConnectionString)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal koneksi ke database: " + err.Error()})
		return
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal ping database: " + err.Error()})
		return
	}

	// update user berdasarkan GUID
	collection := client.Database(MongoDatabaseName).Collection(MongoCollectionName)

	filter := bson.M{"guid": req.Guid}
	update := bson.M{
		"$set": bson.M{
			"username": req.Username,
			"password": string(hashedPassword),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal update user: " + err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(LoginResponse{Message: "User dengan GUID tersebut tidak ditemukan"})
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{Message: "Edit register (PUT) berhasil"})
}

// editRegPatchHandler - PATCH /api3/edit-reg
// Edit data user berdasarkan GUID menggunakan PATCH (partial update)
// Hanya field yang dikirim yang akan diupdate
func editRegPatchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// hanya izinkan PATCH
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Method not allowed, gunakan PATCH"})
		return
	}

	var req EditRegRequest

	// decode JSON dari body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Invalid JSON: " + err.Error()})
		return
	}

	// GUID wajib
	if req.Guid == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "GUID harus diisi"})
		return
	}

	// PATCH: minimal satu field harus diisi
	if req.Username == "" && req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Minimal satu field (username atau password) harus diisi"})
		return
	}

	// siapkan field yang akan diupdate
	updateFields := bson.M{}
	if req.Username != "" {
		updateFields["username"] = req.Username
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal hash password"})
			return
		}
		updateFields["password"] = string(hashedPassword)
	}

	// koneksi ke MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(MongoConnectionString)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal koneksi ke database: " + err.Error()})
		return
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal ping database: " + err.Error()})
		return
	}

	// update user berdasarkan GUID
	collection := client.Database(MongoDatabaseName).Collection(MongoCollectionName)

	filter := bson.M{"guid": req.Guid}
	update := bson.M{"$set": updateFields}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal update user: " + err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(LoginResponse{Message: "User dengan GUID tersebut tidak ditemukan"})
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{Message: "Edit register (PATCH) berhasil"})
}

// delRegHandler - DELETE /api/del-reg
// Hapus data user berdasarkan GUID
func delRegHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// hanya izinkan DELETE
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Method not allowed, gunakan DELETE"})
		return
	}

	var req struct {
		Guid string `json:"guid"`
	}

	// decode JSON dari body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Invalid JSON: " + err.Error()})
		return
	}

	if req.Guid == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Message: "GUID harus diisi"})
		return
	}

	// koneksi ke MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(MongoConnectionString)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal koneksi ke database: " + err.Error()})
		return
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal ping database: " + err.Error()})
		return
	}

	// hapus user berdasarkan GUID
	collection := client.Database(MongoDatabaseName).Collection(MongoCollectionName)

	filter := bson.M{"guid": req.Guid}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Gagal hapus user: " + err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(LoginResponse{Message: "User dengan GUID tersebut tidak ditemukan"})
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{Message: "Hapus user berhasil"})
}

func main() {
	http.HandleFunc("/api/test", testHandler)
	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/addpict", addpictHandler)
	http.HandleFunc("/api1/edit-reg", editRegHandler)
	http.HandleFunc("/api2/edit-reg", editRegPutHandler)
	http.HandleFunc("/api3/edit-reg", editRegPatchHandler)
	http.HandleFunc("/api/del-reg", delRegHandler)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
