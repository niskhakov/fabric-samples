.PHONY: start

start: clean-wallet | run-leveldb cli-init
start-couch: clean-wallet | run-couchdb cli-init

run-leveldb:
	@cd ../first-network && echo y | ./byfn.sh down && echo y | ./byfn.sh up -c mychannel -a -n

run-couchdb:
	@cd ../first-network && echo y | ./byfn.sh down && echo y | ./byfn.sh up -c mychannel -a -n -s couchdb

clean-wallet:
	@cd app/ && rm -rf wallet
	@cd app/ && cp -R wallet-tmpl/ wallet/

stop:
	@cd ../first-network && ./byfn.sh down

copy:
	@docker cp . cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/

prestart-app:
	@cd app/ && nvm use && node enrollAdmin.js && node registerUser.js

cli: copy
	@docker exec -it cli bash

cli-install:
	@docker exec cli bash ./cli_installCC.sh

cli-instantiate:
	@docker exec cli bash ./cli_instantiateCC.sh

cli-init: copy cli-install cli-instantiate

cli-org1:
	@./cli_org1.sh

cli-org2:
	@./cli_org2.sh

cli-batch:
	@./cli_batch.sh

cli-stress-put:
	@./cli-stressPut.sh 100

stress:
	docker exec cli sh -c "./stress_scenario.sh"

cli-stress:
	@./stress_scenario.sh

logs:
	@cd ../first-network && 
	
fabric-build:
	@cd ${GOPATH}/src/github.com/hyperledger/fabric/ && $(MAKE) dist-clean docker-clean native docker docker-thirdparty

fabric-clean:
	@cd ${GOPATH}/src/github.com/hyperledger/fabric/ && $(MAKE) clean


fabric-test:
	@cd ${GOPATH}/src/github.com/hyperledger/fabric/ && $(MAKE) unit-test-clean unit-test