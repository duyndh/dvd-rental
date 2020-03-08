SERVICE_NAME := customer

build-go:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./build/main main.go	
