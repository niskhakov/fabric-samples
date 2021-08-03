.PHONY: start-network 

start: start-network
stop: stop-network

start-network:
	@cd ../first-network && ./byfn.sh up -c mychannel

stop-network:
	@cd ../first-network && ./byfn.sh down

copy:
	@docker cp . cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/

cli: copy
	@docker exec -it cli bash

cli-install:
	@./cli_installCC.sh

cli-instantiate:
	@./cli_instantiateCC.sh

cli-init: cli-install cli-instantiate

cli-org1:
	@./cli_org1.sh

cli-org2:
	@./cli_org2.sh

cli-batch:
	@./cli_batch.sh

cli-stress-put:
	@./cli-stressPut.sh 100

stress-scenario:
	@./stress_scenario.sh

stress: cli-init stress-scenario
	
fabric-build:
	@cd ${GOPATH}/src/github.com/hyperledger/fabric/ && $(MAKE) dist-clean docker-clean native docker docker-thirdparty

fabric-test:
	@cd ${GOPATH}/src/github.com/hyperledger/fabric/ && $(MAKE) unit-test-clean unit-test