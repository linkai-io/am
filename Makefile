protoc:
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/user.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/org.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/address.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/ctserver.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/scangroup/scangroupservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/organization/organizationservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/user/userservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/address/addressservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/coordinator/coordinatorservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/dispatcher/dispatcherservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/module/moduleservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/certtranscoordinator/certtranscoordinatorservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/ctworker/ctworkerservicer.proto


orgservice:
	docker build -t linkai_orgservice -f Dockerfile.orgservice .

userservice:
	docker build -t linkai_userservice -f Dockerfile.userservice .
	
jobservice:
	docker build -t linkai_jobservice -f Dockerfile.jobservice .
	
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

brutemoduleservice:
	docker build -t linkai_brutemoduleservice -f Dockerfile.brutemoduleservice .

services: orgservice userservice scangroupservice addressservice coordinatorservice dispatcherservice amloadservice nsmoduleservice brutemoduleservice

test:
	go test -v ./...
