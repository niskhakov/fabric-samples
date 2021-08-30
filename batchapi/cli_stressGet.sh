#!/bin/bash
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export PEER0_ORG1_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt

VERBOSE="false"
NOBATCHAPI="false"
COLLECTION=""
SEED="1"

EXTRA_ARGS=""

usage() {
  echo "Usage: ${0} [-vn] [-c COLLECTION] [-s SEED_NUM] [-k KEY_LENGTH] NUMBER" >&2
  echo "Queries getManyObjectsBatch chaincode method which gets NUMBER randomly generated keys from the ledger via one batch operation" >&2
  echo "  -v             Verbose mode - chaincode returns generated key/values and parameters" >&2
  echo "  -n             NoBatchAPI mode - for every key will be invoked GetState/GetPrivateData instead of BatchAPI" >&2
  echo "  -s SEED_NUM    Specify seed value to reproduce randomly generated keys" >&2
  echo "  -c COLLECTION  Specify private collection to get key-values" >&2
  echo "  -k KEY_LENGTH  Specify key length to get from the ledger" >&2
}
while getopts s:vnc:k: OPTION; do
  case "${OPTION}" in
    s)
      SEED="${OPTARG}"
      EXTRA_ARGS="${EXTRA_ARGS},\"seed\",\"${SEED}\""
      ;;
    v)
      VERBOSE="true"
      EXTRA_ARGS="${EXTRA_ARGS},\"verbose\""
      ;;
    n)
      NOBATCHAPI="true"
      EXTRA_ARGS="${EXTRA_ARGS},\"nobatchapi\""
      ;;
    c)
      COLLECTION="${OPTARG}"
      EXTRA_ARGS="${EXTRA_ARGS},\"collection\",\"${COLLECTION}\""
      ;;
    k)
      KEYLEN="${OPTARG}"
      EXTRA_ARGS="${EXTRA_ARGS},\"keylength\",\"${KEYLEN}\""
      ;;
    ?)
      usage
      exit 1
  esac
done

shift $((OPTIND - 1))

if [[ "${#}" -lt 1 ]]; then
  usage
  exit 1
fi

NUM=$1

# set -x
peer chaincode query -C mychannel -n batchapicc -c "{\"Args\":[\"getManyObjectsBatch\",\"${NUM}\"${EXTRA_ARGS}]}"
# set +x