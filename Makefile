SERVICE_NAME := customer
IMAGE_NAME := ndhduy7798/dvd_rental_$(SERVICE_NAME)
REDIS_URL := localhost:6379
POSTGRESQL_URL := localhost:5432
POSTGRESQL_USERNAME := password123
ZIPKIN_URL := localhost:9411
SERVICE := customer
NAMESPACE := api

#* Build

build:
	@echo "--> Buildding image"
	@echo $(TESTING)
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
#! Testing
test:
	@echo "--> Testing All Services"
	go test ./...
test-customer:
	@echo "--> Testing Customer"
	go test ./customer/...
#* Publish
publish-latest:
	@echo "Publishing image"
	docker push $(IMAGE_NAME):latest