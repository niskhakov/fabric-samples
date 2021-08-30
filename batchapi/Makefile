########################################
# start - start fabric network, install example chaincode and clean node app artefacts
# start-couch - same as `start`, but with CouchDB instead of LevelDB
# stop - stop fabric network
########################################

.PHONY: start

start: clean-wallet | run-leveldb cli-init
start-couch: clean-wallet | run-couchdb cli-init

# Stop network
stop:
	@cd ../first-network && ./byfn.sh down


# Run network with leveldb
run-leveldb:
	@cd ../first-network && echo y | ./byfn.sh down && echo y | ./byfn.sh up -c mychannel -a -n


# Run network with couchdb
run-couchdb:
	@cd ../first-network && echo y | ./byfn.sh down && echo y | ./byfn.sh up -c mychannel -a -n -s couchdb

# Delete client node app artefacts 
clean-wallet:
	@cd app/ && rm -rf wallet
	@cd app/ && cp -R wallet-tmpl/ wallet/

# Copy scripts to docker fabric cli container
copy:
	@docker cp . cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/


# Go to docker fabric cli container
cli: copy
	@docker exec -it cli bash

# Install chaincode in cli
cli-install:
	@docker exec cli bash ./cli_installCC.sh

# Instantiate chaincode in cli
cli-instantiate:
	@docker exec cli bash ./cli_instantiateCC.sh

cli-init: copy cli-install cli-instantiate

# Run example scenario over Fabric Command Line Interface (should be run in cli container - `make cli`)
cli-batch:
	@./cli_batch.sh


########################################
# Fabric developer instructions 
########################################

fabric-build:
	@cd ${GOPATH}/src/github.com/hyperledger/fabric/ && $(MAKE) dist-clean docker-clean native docker docker-thirdparty

fabric-clean:
	@cd ${GOPATH}/src/github.com/hyperledger/fabric/ && $(MAKE) clean


fabric-test:
	@cd ${GOPATH}/src/github.com/hyperledger/fabric/ && $(MAKE) unit-test-clean unit-test