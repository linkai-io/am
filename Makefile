ALL_SERVICES = orgservice userservice scangroupservice addressservice coordinatorservice dispatcherservice nsmoduleservice webdataservice bigdataservice brutemoduleservice
BACKEND_SERVICES = orgservice userservice scangroupservice addressservice coordinatorservice dispatcherservice webdataservice bigdataservice
MODULE_SERVICES = nsmoduleservice brutemoduleservice bigdatamoduleservice
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
	protoc -I ../protorepo/protocservices/ --gofast_out=plugins=grpc:$$GOPATH/src ../protorepo/protocservices/metrics/load.proto

orgservice:
	docker build -t orgservice -f Dockerfile.orgservice .

userservice:
	docker build -t userservice -f Dockerfile.userservice .
	
scangroupservice:
	docker build -t scangroupservice -f Dockerfile.scangroupservice .

addressservice:
	docker build -t addressservice -f Dockerfile.addressservice .

coordinatorservice:
	docker build -t coordinatorservice -f Dockerfile.coordinatorservice .

dispatcherservice:
	docker build -t dispatcherservice -f Dockerfile.dispatcherservice .

nsmoduleservice:
	docker build -t nsmoduleservice -f Dockerfile.nsmoduleservice .

webdataservice:
	docker build -t webdataservice -f Dockerfile.webdataservice .

bigdataservice:
	docker build -t bigdataservice -f Dockerfile.bigdataservice .

brutemoduleservice:
	docker build -t brutemoduleservice -f Dockerfile.brutemoduleservice .

webmoduleservice:
	docker build -t webmoduleservice -f Dockerfile.webmoduleservice .

bigdatamoduleservice:
	docker build -t bigdatamoduleservice -f Dockerfile.bigdatamoduleservice .

allservices: orgservice userservice scangroupservice addressservice coordinatorservice dispatcherservice nsmoduleservice webdataservice bigdataservice brutemoduleservice webmoduleservice bigdatamoduleservice

backend: orgservice userservice scangroupservice addressservice coordinatorservice dispatcherservice webdataservice bigdataservice

pushbackend: 
	$(foreach var,$(BACKEND_SERVICES),docker tag $(var):latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/$(var):latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/$(var):latest;)

pushallservices:
	$(foreach var,$(ALL_SERVICES),docker tag $(var):latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/$(var):latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/$(var):latest;)

pushnsmoduleservice: nsmoduleservice
	docker tag nsmoduleservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/nsmoduleservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/nsmoduleservice:latest

pushbrutemoduleservice: brutemoduleservice
	docker tag brutemoduleservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/brutemoduleservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/brutemoduleservice:latest

pushbigdatamoduleservice: bigdatamoduleservice
	docker tag bigdatamoduleservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/bigdatamoduleservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/bigdatamoduleservice:latest

pushorgservice: orgservice
	docker tag orgservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/orgservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/orgservice:latest

pushuserservice: userservice
	docker tag userservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/userservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/userservice:latest

pushaddressservice: addressservice
	docker tag addressservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/addressservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/addressservice:latest

pushscangroupservice: scangroupservice 
	docker tag scangroupservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/scangroupservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/scangroupservice:latest

pushcoordinatorservice:
	docker tag coordinatorservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/coordinatorservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/coordinatorservice:latest

pushdispatcherservice:
	docker tag dispatcherservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/dispatcherservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/dispatcherservice:latest

pushwebdataservice:
	docker tag webdataservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/webdataservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/webdataservice:latest

pushbigdataservice:
	docker tag bigdataservice:latest 447064213022.dkr.ecr.us-east-1.amazonaws.com/bigdataservice:latest && docker push 447064213022.dkr.ecr.us-east-1.amazonaws.com/bigdataservice:latest

deploybackend: 
	$(foreach var,$(BACKEND_SERVICES),aws ecs update-service --cluster ${APP_ENV}-backend-ecs-cluster --force-new-deployment --service $(var);)

deploymodules:
	$(foreach var,$(MODULE_SERVICES),aws ecs update-service --cluster ${APP_ENV}-modules-ecs-cluster --force-new-deployment --service $(var);)

deploybrutemoduleservice:
	aws ecs update-service --cluster ${APP_ENV}-modules-ecs-cluster --force-new-deployment --service brutemoduleservice

deploybigdatamoduleservice:
	aws ecs update-service --cluster ${APP_ENV}-modules-ecs-cluster --force-new-deployment --service bigdatamoduleservice	

deploycoordinatorservice:
	aws ecs update-service --cluster ${APP_ENV}-backend-ecs-cluster --force-new-deployment --service coordinatorservice

deploydispatcherservice:
	aws ecs update-service --cluster ${APP_ENV}-backend-ecs-cluster --force-new-deployment --service dispatcherservice

deployscangroupservice: pushscangroupservice 
	aws ecs update-service --cluster ${APP_ENV}-backend-ecs-cluster --force-new-deployment --service scangroupservice

deploywebdataservice:
	aws ecs update-service --cluster ${APP_ENV}-backend-ecs-cluster --force-new-deployment --service webdataservice

deploybigdataservice:
	aws ecs update-service --cluster ${APP_ENV}-backend-ecs-cluster --force-new-deployment --service bigdataservice

pushwebmoduleservice:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o deploy_files/webmoduleservice cmd/module/web/main.go	
	zip deploy_files/webmodule.zip third_party/local.conf deploy_files/webmoduleservice
	aws s3 cp deploy_files/webmodule.zip s3://linkai-infra/${APP_ENV}/webmodule/webmodule.zip

test:
	go test ./... -cover

infratest:
	INFRA_TESTS=yes go test ./... -cover