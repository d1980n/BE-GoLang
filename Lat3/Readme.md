# Install Swagger di Golang (Swaggo)

Panduan ini menjelaskan cara menginstall dan menggunakan Swagger di Golang menggunakan **Swaggo (`swag`)**.

---

## 🚀 1. Install Swagger CLI (`swag`)

Jalankan perintah berikut di terminal:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Cek apakah instalasi berhasil:

```bash
swag --version
```

---

## ⚠️ Jika `swag` Tidak Dikenali

Tambahkan ke PATH:

### Linux / Mac

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Windows (CMD)

```bash
set PATH=%PATH%;%GOPATH%\bin
```

---

## 📦 2. Install Library Swagger di Project

Masuk ke folder project Go:

```bash
go mod init nama_project   # jika belum ada go.mod
```

Install dependency Swagger:

```bash
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
```

---

## 🧪 3. Generate Swagger Docs

Setelah menambahkan komentar Swagger di kode Go, jalankan:

```bash
swag init
```

Jika berhasil, akan muncul folder:

```
docs/
```

---

## 🌐 4. Akses Swagger UI

Setelah server dijalankan, buka di browser:

```
http://localhost:8080/swagger/index.html
```

---

## 🧠 Ringkasan

| Langkah         | Perintah                                            |
| --------------- | --------------------------------------------------- |
| Install CLI     | `go install github.com/swaggo/swag/cmd/swag@latest` |
| Cek versi       | `swag --version`                                    |
| Init module     | `go mod init nama_project`                          |
| Install library | `go get -u github.com/swaggo/gin-swagger`           |
| Generate docs   | `swag init`                                         |
| Jalankan server | `go run .`                                          |

---

## ⚡ Catatan Penting

* `swag` adalah generator → membutuhkan komentar annotation di kode
* Pastikan file `go.mod` sudah ada
* Jalankan ulang `swag init` setiap ada perubahan API
* Gunakan framework seperti Gin untuk integrasi yang lebih mudah

---

Dengan langkah ini, Swagger siap digunakan untuk dokumentasi API di project Golang Anda 🚀
