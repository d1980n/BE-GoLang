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
