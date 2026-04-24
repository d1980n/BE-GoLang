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

