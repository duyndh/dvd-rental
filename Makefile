SERVICE_NAME := customer
IMAGE_NAME := ndhduy7798/dvd_rental_$(SERVICE_NAME)
zipkinAddr := localhost:32770
dbHost := localhost:32768
dbUserName := my_user
dbPassword := dbPassword
redisAddr := localhost:32769
#* Build
build:
	@echo "--> Buildding image"
	docker build . -t $(IMAGE_NAME):latest /
	--build-arg REDIS_URL=$(REDIS_URL) --build-arg POSTGRESQL_URL=$(POSTGRESQL_URL) /
	--build-arg POSTGRESQL_USERNAME=$(POSTGRESQL_USERNAME)/
	--build-arg POSTGRESQL_PASSWORD = $(POSTGRESQL_PASSWORD)/
	--build-arg ZIPKIN_URL=$(ZIPKIN_URL)/
	--build-arg SERVICE=$(SERVICE) /
	--build-arg NAMESPACE=$(NAMESPACE)
build-go:
	@echo "--> Building go"
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./build/main main.go	
run-customer:
	go run main.go -zipkinAddr=${zipkinAddr} -dbHost=${dbHost} -dbUserName={my_user} -dbPassword=${dbPassword} -redisAddr=${redisAddr} -service=customer -namespace=api -grpcAddr=localhost:8888
run-dvd:
	go run main.go -zipkinAddr=${zipkinAddr} -dbHost=${dbHost} -dbUserName={my_user} -dbPassword=${dbPassword} -redisAddr=${redisAddr} -service=dvd -namespace=svc

#! Testing
test:
	@echo "--> Testing All Services"
	go test ./...
test-customer:
	@echo "--> Testing Customer"
	go test ./customer/...
test-dvd:
	@echo "--> Testing Customer"
	go test ./dvd/...

#* Publish
publish-latest:
	@echo "Publishing image"
	docker push $(IMAGE_NAME):latest

#* Docker-compose
up:
	docker-compose up -d --build
down:
	docker-compose down
stop:
	docker-compose stop
start:
	docker-compose start