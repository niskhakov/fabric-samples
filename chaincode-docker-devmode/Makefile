.PHONY:
.SILENT:

DEVMODE=${HOME}/github.com/niskhakov/fabric-samples/chaincode-docker-devmode

docker-rm:
	docker rm -f $(shell docker ps -aq)

docker-rmi:
	docker rmi -f $(shell docker images | grep fabcar | awk '{print $3}')

compose:
	cd $(DEVMODE) && docker-compose -f docker-compose-simple.yaml up

composed: docker-rm compose

build:
	go build

build-cc:
	docker exec -it chaincode bash -c "cd sacc && go build"

start-cc: build-cc
	docker exec -it chaincode bash -c "cd sacc && CORE_PEER_ADDRESS=peer:7052 CORE_CHAINCODE_ID_NAME=mycc:0 ./sacc"

init-install:
	@peer chaincode install -p chaincodedev/chaincode/sacc -n mycc -v 0

init-instantiate:
	@peer chaincode instantiate -n mycc -v 0 -c '{"Args":["a","10"]}' -C myc

init: init-install init-instantiate

one-invoke:
	@peer chaincode invoke -n mycc -c '{"Args":["set","argB","2"]}' -C myc

one-query:
	@peer chaincode query -n mycc -c '{"Args":["get","argB"]}' -C myc

multi-invoke:
	@peer chaincode invoke -n mycc -c '{"Args":["multiset","one","one","two","two","argD","4","argE","50"]}' -C myc

multi-query:
	@peer chaincode query -n mycc -c '{"Args":["multiget","a","one","two","argB","argC","argD","argE"]}' -C myc

cli:
	docker exec -it cli bash

fabric-build:
	@cd ${GOPATH}/src/github.com/hyperledger/fabric/ && $(MAKE) dist-clean docker-clean native docker docker-thirdparty
