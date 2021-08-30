# Installing Chaincode
peer chaincode install -n batchapicc -v 1.0 -p github.com/hyperledger/fabric/peer/chaincode/go/

export CORE_PEER_ADDRESS=peer1.org1.example.com:8051
peer chaincode install -n batchapicc -v 1.0 -p github.com/hyperledger/fabric/peer/chaincode/go/

export CORE_PEER_LOCALMSPID=Org2MSP
export PEER0_ORG2_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG2_CA
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp

export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
peer chaincode install -n batchapicc -v 1.0 -p github.com/hyperledger/fabric/peer/chaincode/go/

export CORE_PEER_ADDRESS=peer1.org2.example.com:10051
peer chaincode install -n batchapicc -v 1.0 -p github.com/hyperledger/fabric/peer/chaincode/go/
