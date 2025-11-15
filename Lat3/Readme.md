# Cross Compile (langsung)
## di linux
Cek Arch: uname -m / lscpu
## di windows
Build Go: GOOS=linux GOARCH=${ARCH-TYPE} go build -o bin/main
arm32: GOOS=linux GOARCH=arm GOARM=7 go build -o bin/main
