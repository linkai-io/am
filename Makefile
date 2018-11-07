build:
	go build -v ./...

protoc:
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/user.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/org.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/address.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/web.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/ctrecord.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/scangroup/scangroupservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/organization/organizationservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/user/userservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/address/addressservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/coordinator/coordinatorservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/dispatcher/dispatcherservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/module/moduleservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/webdata/webdataservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/bigdata/bigdataservicer.proto

orgservice:
	docker build -t linkai_orgservice -f Dockerfile.orgservice .

userservice:
	docker build -t linkai_userservice -f Dockerfile.userservice .
	
scangroupservice:
	docker build -t linkai_scangroupservice -f Dockerfile.scangroupservice .

addressservice:
	docker build -t linkai_addressservice -f Dockerfile.addressservice .

coordinatorservice:
	docker build -t linkai_coordinatorservice -f Dockerfile.coordinatorservice .

dispatcherservice:
	docker build -t linkai_dispatcherservice -f Dockerfile.dispatcherservice .

nsmoduleservice:
	docker build -t linkai_nsmoduleservice -f Dockerfile.nsmoduleservice .

amloadservice:
	docker build -t linkai_amloadservice -f Dockerfile.amloadservice .

webdataservice:
	docker build -t linkai_webdataservice -f Dockerfile.webdataservice .

brutemoduleservice:
	docker build -t linkai_brutemoduleservice -f Dockerfile.brutemoduleservice .

webmoduleservice:
	docker build -t linkai_webmoduleservice -f Dockerfile.webmoduleservice .

services: orgservice userservice scangroupservice addressservice coordinatorservice dispatcherservice nsmoduleservice amloadservice webdataservice brutemoduleservice webmoduleservice

test:
	go test ./... -cover

infratest:
	INFRA_TESTS=yes go test ./... -cover