package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "doc7/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Example API
// @version 1.0
// @description This is a sample server for Swagger documentation.
// @host localhost:8082
// @BasePath /
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerHome)
	mux.HandleFunc("/api/test", handlerApiTest)
	//api/1var?nilai=10
	mux.HandleFunc("/api/1var", handlerApi1Var)
	//api/1varRESTFUL/10
	mux.HandleFunc("/api/1varRESTFUL/", handlerApi1VarRESTFUL)
	//api/2var?nilai1=10&nilai2=20
	mux.HandleFunc("/api/2var", handlerApi2Var)
	//api/2varRESTFUL/10/20
	mux.HandleFunc("/api/2varRESTFUL/", handlerApi2VarRESTFUL)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// 3. Konfigurasi dan Jalankan Server
	port := ":8082"
	log.Printf("Server siap berjalan di http://localhost%s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// handlerHome godoc
// @Summary Home endpoint
// @Description Show welcome message
// @Tags home
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Internal Server Error"
// @Router / [get]
func handlerHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte("Welcome to the API!\n"))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type ResponTest struct {
	Nilai int `json:"nilai"`
}

// handlerApiTest godoc
// @Summary Test API
// @Description Get a test value
// @Tags test
// @Success 200 {object} ResponTest "Success"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/test [get]
func handlerApiTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := ResponTest{Nilai: 2}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handlerApi1Var godoc
// @Summary Get nilai from query
// @Description Get nilai from query parameter, default 10 if not set
// @Tags test
// @Param nilai query int false "Nilai"
// @Success 200 {object} ResponTest "Success"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/1var [get]
func handlerApi1Var(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Ambil query parameter
	nilaiStr := r.URL.Query().Get("nilai")

	var nilai int
	if nilaiStr != "" {
		fmt.Sscanf(nilaiStr, "%d", &nilai)
	} else {
		nilai = 10 // default
	}

	data := ResponTest{
		Nilai: nilai,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handlerApi1VarRESTFUL godoc
// @Summary Get nilai from URL path (RESTful style)
// @Description Get nilai from the URL path parameter, e.g. /api/1varRESTFUL/10
// @Tags test
// @Produce json
// @Param nilai path int true "Nilai parameter"
// @Success 200 {object} ResponTest "Success"
// @Failure 400 {string} string "Bad Request"
// @Failure 405 {string} string "Method Not Allowed"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/1varRESTFUL/{nilai} [get]
func handlerApi1VarRESTFUL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Ambil bagian path setelah prefix, contoh: /api/1varRESTFUL/10 -> "10"
	nilaiStr := strings.TrimPrefix(r.URL.Path, "/api/1varRESTFUL/")

	var nilai int
	if nilaiStr == "" {
		http.Error(w, "Parameter nilai diperlukan, contoh: /api/1varRESTFUL/10", http.StatusBadRequest)
		return
	}
	_, err := fmt.Sscanf(nilaiStr, "%d", &nilai)
	if err != nil {
		http.Error(w, "Parameter nilai harus berupa angka", http.StatusBadRequest)
		return
	}

	data := ResponTest{Nilai: nilai}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handlerApi2Var(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Ambil query parameter pasti string
	nilai1Str := r.URL.Query().Get("nilai1")
	nilai2Str := r.URL.Query().Get("nilai2")

	var n1 int
	var n2 int

	// Convert string ke int
	fmt.Sscanf(nilai1Str, "%d", &n1)
	fmt.Sscanf(nilai2Str, "%d", &n2)

	// Contoh: jumlahkan
	hasil := n1 + n2

	response := map[string]int{
		"nilai1": n1,
		"nilai2": n2,
		"hasil":  hasil,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handlerApi2VarRESTFUL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Ambil bagian path setelah prefix, contoh: /api/2varRESTFUL/3/5 -> "3/5"
	path := r.URL.Path
	//nilaiStr := strings.TrimPrefix(path, "/api/2varRESTFUL/")

	var n1, n2 int
	_, err := fmt.Sscanf(path, "/api/2varRESTFUL/%d/%d", &n1, &n2)

	if err != nil {
		http.Error(w, "Parameter nilai diperlukan, contoh: /api/2varRESTFUL/3/5", http.StatusBadRequest)
		return
	}

	hasil := n1 + n2

	response := map[string]int{
		"nilai1": n1,
		"nilai2": n2,
		"hasil":  hasil,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
