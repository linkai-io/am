ALL_SERVICES = orgservice userservice scangroupservice addressservice coordinatorservice dispatcherservice nsmoduleservice webdataservice brutemoduleservice
BACKEND_SERVICES = orgservice userservice scangroupservice addressservice coordinatorservice dispatcherservice webdataservice
APP_ENV = dev
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

allservices: orgservice userservice scangroupservice addressservice coordinatorservice dispatcherservice nsmoduleservice amloadservice webdataservice brutemoduleservice webmoduleservice

backend: orgservice userservice scangroupservice addressservice coordinatorservice dispatcherservice webdataservice

pushbackend: 
	$(foreach var,$(BACKEND_SERVICES),docker tag linkai_$(var):latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/$(var):latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/$(var):latest;)

pushecr:
	$(foreach var,$(ALL_SERVICES),docker tag linkai_$(var):latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/$(var):latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/$(var):latest;)

pushorgservice: orgservice
	docker tag linkai_orgservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/orgservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/orgservice:latest

pushuserservice: userservice
	docker tag linkai_userservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/userservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/userservice:latest



deploy_loadbalancer:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o deploy_files/loadbalancer cmd/amload/main.go
	zip deploy_files/loadbalancer.zip deploy_files/loadbalancer deploy_files/30-loadbalancer.conf deploy_files/install_loadbalancer.sh deploy_files/loadbalancer.service
	aws s3 cp deploy_files/loadbalancer.zip s3://linkai-infra/${APP_ENV}/loadbalancer/loadbalancer.zip

deploy_webmoduleservice:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o deploy_files/webmoduleservice cmd/module/web/main.go	
	zip deploy_files/webmodule.zip third_party/local.conf deploy_files/webmoduleservice
	aws s3 cp deploy_files/webmodule.zip s3://linkai-infra/${APP_ENV}/webmodule/webmodule.zip

test:
	go test ./... -cover

infratest:
	INFRA_TESTS=yes go test ./... -cover