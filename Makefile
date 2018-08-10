protoc:
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/user.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/prototypes/org.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/scangroup/scangroupservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/organization/organizationservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/user/userservicer.proto
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/address/addressservicer.proto

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

services: orgservice userservice scangroupservice addressservice

test:
	go test -v ./...
