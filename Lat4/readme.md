Contoh coding --> 
| Operasi | Method      |
| ------- | ----------- |
| Create  | POST        |
| Read    | GET         |
| Update  | PUT / PATCH |
| Delete  | DELETE      |


# POST
Cara Test (POST)
🔸 Pakai curl
curl -X POST http://localhost:8080/api/login \
-H "Content-Type: application/json" \
-d '{"username":"admin","password":"1234"}'
🔸 Pakai Postman
Method: POST
URL: http://localhost:8080/api/login
Body → raw → JSON:
{
  "username": "admin",
  "password": "1234"
}

Perbedaan dengan POST dan PUT: PATCH hanya update field yang dikirim (partial update).

Endpoint	Method	Validasi
/api1/edit-reg	POST	guid, username, password semua wajib
/api2/edit-reg	PUT	guid, username, password semua wajib
/api3/edit-reg	PATCH	guid wajib, username/password opsional (minimal satu)
Contoh PATCH — hanya update username:

json
{"guid":"...", "username":"baru_saja"}
Contoh PATCH — hanya update password:

json
{"guid":"...", "password":"pass_baru"}

Build successful ✅.

Endpoint baru: PUT /api2/edit-reg
Sama persis dengan /api1/edit-reg tapi menggunakan method PUT bukan POST.

Test:

PUT http://localhost:8080/api2/edit-reg
Body JSON: {"guid":"...", "username":"baru", "password":"baru123"}
Endpoint	Method	Fungsi
/api1/edit-reg	POST	Edit user by GUID
/api2/edit-reg	PUT	Edit user by GUID
