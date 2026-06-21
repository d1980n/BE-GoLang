##Install Monggo DB dulu

go get go.mongodb.org/mongo-driver/mongo

go get go.mongodb.org/mongo-driver/bson


##compile agar bisa digunakan di linux (dari CMD)
 1. Atur OS target ke Linux
>> $env:GOOS="linux"
>>
>> # 2. Atur arsitektur target ke AMD64
>> $env:GOARCH="amd64"
>>
>> # 3. Lakukan build
>> go build -o go-api-server main3.go
>>
>> 

##Compile (dari linux)
chmod +x go-api-server
./go-api-server

# 4. file: main4.go
>> tambahan /picture masih dalam progres
# 5. file: main5.go
>> perbaikan /picture/CAM2-0001.jpg
>>
# Cross Compile (langsung)
## di linux
``` sh
Cek Arch: uname -m / lscpu
```
## di windows
``` sh
Linux AMD64 (Standar/Server x86_64): api-tera/build/api-tera-linux-amd64

```
============================================================
# Menjalankan Beberapa File `.go` dalam Satu Proyek

Jika Anda memiliki beberapa file `.go` dalam satu proyek dan ingin
menjalankannya menggunakan terminal, ada beberapa cara untuk
melakukannya:

------------------------------------------------------------------------

## **1. Jalankan File Secara Langsung (Tanpa Compile)**

Jika proyek Anda memiliki beberapa file `.go` yang saling terkait dalam
satu package, cukup jalankan perintah berikut di dalam direktori proyek:

``` sh
go run main.go file1.go file2.go
```

Perintah ini akan menjalankan semua file `.go` yang disebutkan secara
langsung.

------------------------------------------------------------------------

## **2. Jalankan Seluruh File dalam Satu Folder**

Jika semua file `.go` ada dalam satu folder dan menggunakan package yang
sama, Anda bisa menjalankannya dengan:

``` sh
go run .
```

atau:

``` sh
go run *.go
```

------------------------------------------------------------------------

## **3. Compile dan Jalankan**

Jika ingin mengompilasi semua file `.go` menjadi satu file biner yang
dapat dieksekusi:

``` sh
go build -o myapp
```

Lalu jalankan hasilnya:

``` sh
./myapp
```

------------------------------------------------------------------------

## **4. Menggunakan Modul Go**

Jika proyek Anda menggunakan **Go Modules**, pastikan file `go.mod`
sudah diinisialisasi dengan:

``` sh
go mod init nama_proyek
```

Lalu gunakan:

``` sh
go run .
```

------------------------------------------------------------------------

## **Catatan**

-   Semua file `.go` harus berada dalam package yang sama (misalnya
    `package main` jika ingin langsung dieksekusi).
-   Jika ada beberapa package dalam satu proyek, pastikan package utama
    (`main`) memanggil fungsi dari package lain dengan `import`.

------------------------------------------------------------------------

Apakah Anda mengalami error saat menjalankan beberapa file `.go`? 🚀
