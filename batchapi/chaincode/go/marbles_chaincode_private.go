/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke marbles, pass private data as base64 encoded bytes in transient map ====
//
// export MARBLE=$(echo -n "{\"name\":\"marble1\",\"color\":\"blue\",\"size\":35,\"owner\":\"tom\",\"price\":99}" | base64 | tr -d \\n)
// peer chaincode invoke -C mychannel -n marblesp -c '{"Args":["initMarble"]}' --transient "{\"marble\":\"$MARBLE\"}"
//
// export MARBLE=$(echo -n "{\"name\":\"marble2\",\"color\":\"red\",\"size\":50,\"owner\":\"tom\",\"price\":102}" | base64 | tr -d \\n)
// peer chaincode invoke -C mychannel -n marblesp -c '{"Args":["initMarble"]}' --transient "{\"marble\":\"$MARBLE\"}"
//
// export MARBLE=$(echo -n "{\"name\":\"marble3\",\"color\":\"blue\",\"size\":70,\"owner\":\"tom\",\"price\":103}" | base64 | tr -d \\n)
// peer chaincode invoke -C mychannel -n marblesp -c '{"Args":["initMarble"]}' --transient "{\"marble\":\"$MARBLE\"}"
//
// export MARBLE_OWNER=$(echo -n "{\"name\":\"marble2\",\"owner\":\"jerry\"}" | base64 | tr -d \\n)
// peer chaincode invoke -C mychannel -n marblesp -c '{"Args":["transferMarble"]}' --transient "{\"marble_owner\":\"$MARBLE_OWNER\"}"
//
// export MARBLE_DELETE=$(echo -n "{\"name\":\"marble1\"}" | base64 | tr -d \\n)
// peer chaincode invoke -C mychannel -n marblesp -c '{"Args":["delete"]}' --transient "{\"marble_delete\":\"$MARBLE_DELETE\"}"

// ==== Query marbles, since queries are not recorded on chain we don't need to hide private data in transient map ====
// peer chaincode query -C mychannel -n marblesp -c '{"Args":["readMarble","marble1"]}'
// peer chaincode query -C mychannel -n marblesp -c '{"Args":["readMarblePrivateDetails","marble1"]}'
// peer chaincode query -C mychannel -n marblesp -c '{"Args":["getMarblesByRange","marble1","marble4"]}'
//
// Rich Query (Only supported if CouchDB is used as state database):
//   peer chaincode query -C mychannel -n marblesp -c '{"Args":["queryMarblesByOwner","tom"]}'
//   peer chaincode query -C mychannel -n marblesp -c '{"Args":["queryMarbles","{\"selector\":{\"owner\":\"tom\"}}"]}'

// INDEXES TO SUPPORT COUCHDB RICH QUERIES
//
// Indexes in CouchDB are required in order to make JSON queries efficient and are required for
// any JSON query with a sort. As of Hyperledger Fabric 1.1, indexes may be packaged alongside
// chaincode in a META-INF/statedb/couchdb/indexes directory. Or for indexes on private data
// collections, in a META-INF/statedb/couchdb/collections/<collection_name>/indexes directory.
// Each index must be defined in its own text file with extension *.json with the index
// definition formatted in JSON following the CouchDB index JSON syntax as documented at:
// http://docs.couchdb.org/en/2.1.1/api/database/find.html#db-index
//
// This marbles02_private example chaincode demonstrates a packaged index which you
// can find in META-INF/statedb/couchdb/collection/collectionMarbles/indexes/indexOwner.json.
// For deployment of chaincode to production environments, it is recommended
// to define any indexes alongside chaincode so that the chaincode and supporting indexes
// are deployed automatically as a unit, once the chaincode has been installed on a peer and
// instantiated on a channel. See Hyperledger Fabric documentation for more details.
//
// If you have access to the your peer's CouchDB state database in a development environment,
// you may want to iteratively test various indexes in support of your chaincode queries.  You
// can use the CouchDB Fauxton interface or a command line curl utility to create and update
// indexes. Then once you finalize an index, include the index definition alongside your
// chaincode in the META-INF/statedb/couchdb/indexes directory or
// META-INF/statedb/couchdb/collections/<collection_name>/indexes directory, for packaging
// and deployment to managed environments.
//
// In the examples below you can find index definitions that support marbles02_private
// chaincode queries, along with the syntax that you can use in development environments
// to create the indexes in the CouchDB Fauxton interface.
//

//Example hostname:port configurations to access CouchDB.
//
//To access CouchDB docker container from within another docker container or from vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for docType, owner.
// Note that docType and owner fields must be prefixed with the "data" wrapper
//
// Index definition for use with Fauxton interface
// {"index":{"fields":["data.docType","data.owner"]},"ddoc":"indexOwnerDoc", "name":"indexOwner","type":"json"}

// Index for docType, owner, size (descending order).
// Note that docType, owner and size fields must be prefixed with the "data" wrapper
//
// Index definition for use with Fauxton interface
// {"index":{"fields":[{"data.size":"desc"},{"data.docType":"desc"},{"data.owner":"desc"}]},"ddoc":"indexSizeSortDoc", "name":"indexSizeSortDesc","type":"json"}

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C mychannel -n marblesp -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":\"marble\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C mychannel -n marblesp -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":{\"$eq\":\"marble\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

const (
	defaultSeed = 1
	keyLength   = 7
)

// Isolate specified rand seed only to methods which use `seededRand`
var seededRand *rand.Rand = rand.New(
	rand.NewSource(defaultSeed))

// Reset random seed of `seededRand` object which is used in RandStringWithCharset and RandString
func RandReset(seed int) {
	seededRand.Seed(int64(seed))
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandString(length int) string {
	return RandStringWithCharset(length, charset)
}

type marble struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Color      string `json:"color"`
	Size       int    `json:"size"`
	Owner      string `json:"owner"`
}

type marblePrivateDetails struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Price      int    `json:"price"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	switch function {
	case "initMarble":
		//create a new marble
		return t.initMarble(stub, args)
	case "readMarble":
		//read a marble
		return t.readMarble(stub, args)
	case "readMarblePrivateDetails":
		//read a marble private details
		return t.readMarblePrivateDetails(stub, args)
	case "transferMarble":
		//change owner of a specific marble
		return t.transferMarble(stub, args)
	case "delete":
		//delete a marble
		return t.delete(stub, args)
	case "queryMarblesByOwner":
		//find marbles for owner X using rich query
		return t.queryMarblesByOwner(stub, args)
	case "queryMarbles":
		//find marbles based on an ad hoc rich query
		return t.queryMarbles(stub, args)
	case "getMarblesByRange":
		//get marbles based on range query
		return t.getMarblesByRange(stub, args)
	case "getMarblesBatch":
		//get multiple marbles via one request
		return t.getMarblesBatch(stub, args)
	case "getManyMarblesBatch":
		//get multiple randomly selected marbles via one request
		return t.getManyMarblesBatch(stub, args)
	case "putMarblesBatch":
		//put multiple marbles via one request
		return t.putMarblesBatch(stub, args)
	case "putManyMarblesBatch":
		// stress test putting multiple marbles via one request
		return t.putManyMarblesBatch(stub, args)
	case "delManyMarblesBatch":
		// stress test deleting multiple marbles via one request
		return t.delManyMarblesBatch(stub, args)
	default:
		//error
		fmt.Println("invoke did not find func: " + function)
		return shim.Error("Received unknown function invocation")
	}
}

// ============================================================
// putMarblesBatch - put marbles info via one network request
// ============================================================
func (t *SimpleChaincode) putMarblesBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error(fmt.Errorf("Incorrect arguments. Expecting at least a key and a value").Error())
	}

	if len(args)%2 != 0 {
		return shim.Error(fmt.Errorf("Incorrect arguments. Expecting even number of arguments: k1, v1, k2, v2, ..., kn, vn").Error())
	}

	kvMap := make([]shim.StateKV, 0)
	for i := 0; i < len(args); i += 2 {
		k := args[i]
		v := args[i+1]
		kvMap = append(kvMap, shim.StateKV{Collection: "", Key: k, Value: []byte(v)})
	}

	err := stub.PutStateBatch(kvMap)
	if err != nil {
		return shim.Error(fmt.Errorf("Failet to set multiple assets: %v with error: %w", kvMap, err).Error())
	}

	// Buffer should be used
	res := ""
	for _, kv := range kvMap {
		res += fmt.Sprintf("%s: %s (%s)\n", kv.Key, kv.Value, kv.Collection)
	}

	return shim.Success([]byte(res))
}

// ============================================================
// putManyMarblesBatch - stress test putting random (with seed) marbles info via one network request
// ============================================================
func (t *SimpleChaincode) putManyMarblesBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error(fmt.Errorf("Incorrect arguments. Expecting at least one argument - number of random keys/value to write in the ledger").Error())
	}

	keyQty, err := strconv.Atoi(args[0]) // number of random kv to write - required
	if err != nil {
		return shim.Error(err.Error())
	}

	var verboseFlag bool
	var useBatchAPI bool = true
	var seedParam int
	var collectionParam string

	// check for verbose param
	if find(args, "verbose") != -1 {
		verboseFlag = true
	}

	// check for nobatchapi param
	if indx := find(args, "nobatchapi"); indx != -1 {
		useBatchAPI = false
	}

	// check for seed param
	if indx := find(args, "seed"); indx != -1 && indx+1 < len(args) {
		seedParam, err = strconv.Atoi(args[indx+1])
		if err != nil {
			seedParam = defaultSeed
		}
	} else {
		seedParam = defaultSeed
	}

	// check for collection param
	if indx := find(args, "collection"); indx != -1 && indx+1 < len(args) {
		collectionParam = args[indx+1]
	}

	RandReset(seedParam)
	keys := make([]string, 0)
	kvMap := make([]shim.StateKV, 0)
	for i := 0; i < keyQty; i++ {
		k := RandString(keyLength)
		keys = append(keys, k)
		v := RandString(keyLength)
		collection := collectionParam
		kvMap = append(kvMap, shim.StateKV{Collection: collection, Key: k, Value: []byte(v)})
	}

	var start time.Time
	var duration time.Duration
	if useBatchAPI {
		// BatchAPI used
		start = time.Now()
		err = stub.PutStateBatch(kvMap)
		duration = time.Since(start)
	} else {
		// BatchAPI is not used, query standard PutState/PutPrivateData for every key
		// use `if` here: to determine whether data is private or not
		if collectionParam != "" {

			start = time.Now()
			for _, kv := range kvMap {
				err = stub.PutPrivateData(collectionParam, kv.Key, kv.Value)
				if err != nil {
					break
				}
			}
			duration = time.Since(start)

		} else {

			start = time.Now()
			for _, kv := range kvMap {
				err = stub.PutState(kv.Key, kv.Value)
				if err != nil {
					break
				}
			}
			duration = time.Since(start)
		}
	}

	if err != nil {
		return shim.Error(fmt.Errorf("Failet to set multiple assets: %v with error: %w", kvMap, err).Error())
	}

	var verboseMsg string
	if verboseFlag {
		verboseMsg = fmt.Sprintf("useBatchAPI: %t, Collection: `%s`, Seed: %d, Keys: %s", useBatchAPI, collectionParam, seedParam, strings.Join(keys, ", "))
	}

	res := fmt.Sprintf("Put state invoked: putting %d entries in the ledger takes %s %s", keyQty, duration.String(), verboseMsg)

	return shim.Success([]byte(res))
}

// ============================================================
// getMarblesBatch - get marbles via one network request
// ============================================================
func (t *SimpleChaincode) getMarblesBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error(fmt.Errorf("Incorrect arguments. Expecting at least one key").Error())
	}

	keys := make([]shim.StateKey, 0)
	for _, k := range args {
		keys = append(keys, shim.StateKey{Collection: "", Key: k})
	}
	value, err := stub.GetStateBatch(keys)
	if err != nil {
		return shim.Error(fmt.Errorf("Failed to get asset: %s with error: %s", args, err).Error())
	}
	if len(value) == 0 {
		return shim.Error(fmt.Errorf("Assets not found: %s", args).Error())
	}

	// Buffer should be used
	res := ""
	for _, kv := range value {
		res += fmt.Sprintf("%s: %s (%s)\n", kv.Key, kv.Value, kv.Collection)
	}

	return shim.Success([]byte(res))
}

// ============================================================
// getManyMarblesBatch - get many randomly selected (with seed) marbles via one network request
// ============================================================
func (t *SimpleChaincode) getManyMarblesBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error(fmt.Errorf("Incorrect arguments. Expecting at least one argument").Error())
	}

	keyQty, err := strconv.Atoi(args[0]) // number of random keys to read - required
	if err != nil {
		return shim.Error(err.Error())
	}

	var verboseFlag bool
	var useBatchAPI bool = true
	var seedParam int
	var collectionParam string

	// check for verbose param
	if find(args, "verbose") != -1 {
		verboseFlag = true
	}

	// check for nobatchapi param
	if indx := find(args, "nobatchapi"); indx != -1 {
		useBatchAPI = false
	}

	// check for seed param
	if indx := find(args, "seed"); indx != -1 && indx+1 < len(args) {
		seedParam, err = strconv.Atoi(args[indx+1])
		if err != nil {
			seedParam = defaultSeed
		}
	} else {
		seedParam = defaultSeed
	}

	// check for collection param
	if indx := find(args, "collection"); indx != -1 && indx+1 < len(args) {
		collectionParam = args[indx+1]
	}

	RandReset(seedParam)

	keys := make([]shim.StateKey, 0)
	for i := 0; i < keyQty; i++ {
		keys = append(keys, shim.StateKey{Collection: collectionParam, Key: RandString(keyLength)})

		// Use RandString one more time to be consistent with putManyMarbles, which invokes RandString 2 times
		// and get the same keys as were written in put operation
		_ = RandString(keyLength)
	}

	var start time.Time
	var duration time.Duration
	var value []shim.StateKV
	if useBatchAPI {
		// BatchAPI used
		start = time.Now()
		value, err = stub.GetStateBatch(keys)
		duration = time.Since(start)
	} else {
		// BatchAPI is not used, query standard GetState/GetPrivateData for every key
		value = make([]shim.StateKV, 0, len(keys))
		var singleVal []byte
		// use `if` here: to determine whether data is private or not
		if collectionParam != "" {

			start = time.Now()
			for _, k := range keys {
				singleVal, err = stub.GetPrivateData(collectionParam, k.Key)
				if err != nil {
					break
				}
				value = append(value, shim.StateKV{Key: k.Key, Value: singleVal, Collection: collectionParam})
			}
			duration = time.Since(start)

		} else {

			start = time.Now()
			for _, k := range keys {

				singleVal, err = stub.GetState(k.Key)
				if err != nil {
					break
				}
				value = append(value, shim.StateKV{Key: k.Key, Value: singleVal})
			}
			duration = time.Since(start)
		}
	}

	if err != nil {
		return shim.Error(fmt.Errorf("Failed to get asset: %s with error: %s", args, err).Error())
	}
	if len(value) == 0 {
		return shim.Error(fmt.Errorf("Assets not found: %s", args).Error())
	}

	// Buffer should be used
	var verboseMsg string

	if verboseFlag {
		for _, kv := range value {
			verboseMsg += fmt.Sprintf("%s: %s (collection:`%s`)\n", kv.Key, kv.Value, kv.Collection)
		}
		verboseMsg += fmt.Sprintf("useBatchAPI: %t, Seed: %d", useBatchAPI, seedParam)
	}

	res := fmt.Sprintf("Get state queried: getting %d entries from the ledger takes %s %s", keyQty, duration.String(), verboseMsg)

	return shim.Success([]byte(res))
}

// ============================================================
// delManyMarblesBatch - deletes many randomly selected marbles (with seed) via one network request
// ============================================================
func (t *SimpleChaincode) delManyMarblesBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error(fmt.Errorf("Incorrect arguments. Expecting at least one argument").Error())
	}

	keyQty, err := strconv.Atoi(args[0]) // number of random keys to read - required
	if err != nil {
		return shim.Error(err.Error())
	}

	var verboseFlag bool
	var useBatchAPI bool = true
	var seedParam int
	var collectionParam string

	// check for verbose param
	if find(args, "verbose") != -1 {
		verboseFlag = true
	}

	// check for nobatchapi param
	if indx := find(args, "nobatchapi"); indx != -1 {
		useBatchAPI = false
	}

	// check for seed param
	if indx := find(args, "seed"); indx != -1 && indx+1 < len(args) {
		seedParam, err = strconv.Atoi(args[indx+1])
		if err != nil {
			seedParam = defaultSeed
		}
	} else {
		seedParam = defaultSeed
	}

	// check for collection param
	if indx := find(args, "collection"); indx != -1 && indx+1 < len(args) {
		collectionParam = args[indx+1]
	}

	RandReset(seedParam)

	keys := make([]shim.StateKey, 0)
	for i := 0; i < keyQty; i++ {
		keys = append(keys, shim.StateKey{Collection: collectionParam, Key: RandString(keyLength)})

		// Use RandString one more time to be consistent with putManyMarbles, which invokes RandString 2 times
		// and get the same keys as were written in put operation
		_ = RandString(keyLength)
	}

	var start time.Time
	var duration time.Duration
	if useBatchAPI {
		// BatchAPI used
		start = time.Now()
		err = stub.DelStateBatch(keys)
		duration = time.Since(start)
	} else {
		// BatchAPI is not used, query standard DelState/DelPrivateData for every key
		// use `if` here: to determine whether data is private or not
		if collectionParam != "" {

			start = time.Now()
			for _, k := range keys {
				err = stub.DelPrivateData(collectionParam, k.Key)
				if err != nil {
					break
				}
			}
			duration = time.Since(start)

		} else {

			start = time.Now()
			for _, k := range keys {
				err = stub.DelState(k.Key)
				if err != nil {
					break
				}
			}
			duration = time.Since(start)
		}
	}

	if err != nil {
		return shim.Error(fmt.Errorf("Failed to get asset: %s with error: %s", args, err).Error())
	}

	// Buffer should be used
	var verboseMsg string

	if verboseFlag {
		keysStr := make([]string, 0, len(keys))
		for _, v := range keys {
			keysStr = append(keysStr, v.Key)
		}
		verboseMsg = fmt.Sprintf("useBatchAPI: %t, Collection: `%s`, Seed: %d, Keys: %s", useBatchAPI, collectionParam, seedParam, strings.Join(keysStr, ", "))
	}

	res := fmt.Sprintf("Del state invoked: deleting %d entries from the ledger takes %s %s", keyQty, duration.String(), verboseMsg)

	return shim.Success([]byte(res))
}

// ============================================================
// initMarble - create a new marble, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	type marbleTransientInput struct {
		Name  string `json:"name"` //the fieldtags are needed to keep case from bouncing around
		Color string `json:"color"`
		Size  int    `json:"size"`
		Owner string `json:"owner"`
		Price int    `json:"price"`
	}

	// ==== Input sanitation ====
	fmt.Println("- start init marble")

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private marble data must be passed in transient map.")
	}

	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["marble"]; !ok {
		return shim.Error("marble must be a key in the transient map")
	}

	if len(transMap["marble"]) == 0 {
		return shim.Error("marble value in the transient map must be a non-empty JSON string")
	}

	var marbleInput marbleTransientInput
	err = json.Unmarshal(transMap["marble"], &marbleInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(transMap["marble"]))
	}

	if len(marbleInput.Name) == 0 {
		return shim.Error("name field must be a non-empty string")
	}
	if len(marbleInput.Color) == 0 {
		return shim.Error("color field must be a non-empty string")
	}
	if marbleInput.Size <= 0 {
		return shim.Error("size field must be a positive integer")
	}
	if len(marbleInput.Owner) == 0 {
		return shim.Error("owner field must be a non-empty string")
	}
	if marbleInput.Price <= 0 {
		return shim.Error("price field must be a positive integer")
	}

	// ==== Check if marble already exists ====
	marbleAsBytes, err := stub.GetPrivateData("collectionMarbles", marbleInput.Name)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if marbleAsBytes != nil {
		fmt.Println("This marble already exists: " + marbleInput.Name)
		return shim.Error("This marble already exists: " + marbleInput.Name)
	}

	// ==== Create marble object, marshal to JSON, and save to state ====
	marble := &marble{
		ObjectType: "marble",
		Name:       marbleInput.Name,
		Color:      marbleInput.Color,
		Size:       marbleInput.Size,
		Owner:      marbleInput.Owner,
	}
	marbleJSONasBytes, err := json.Marshal(marble)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save marble to state ===
	err = stub.PutPrivateData("collectionMarbles", marbleInput.Name, marbleJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Create marble private details object with price, marshal to JSON, and save to state ====
	marblePrivateDetails := &marblePrivateDetails{
		ObjectType: "marblePrivateDetails",
		Name:       marbleInput.Name,
		Price:      marbleInput.Price,
	}
	marblePrivateDetailsBytes, err := json.Marshal(marblePrivateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutPrivateData("collectionMarblePrivateDetails", marbleInput.Name, marblePrivateDetailsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marble.Color, marble.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutPrivateData("collectionMarbles", colorNameIndexKey, value)

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init marble")
	return shim.Success(nil)
}

// ===============================================
// readMarble - read a marble from chaincode state
// ===============================================
func (t *SimpleChaincode) readMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetPrivateData("collectionMarbles", name) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ===============================================
// readMarblereadMarblePrivateDetails - read a marble private details from chaincode state
// ===============================================
func (t *SimpleChaincode) readMarblePrivateDetails(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetPrivateData("collectionMarblePrivateDetails", name) //get the marble private details from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get private details for " + name + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble private details does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a marble key/value pair from state
// ==================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start delete marble")

	type marbleDeleteTransientInput struct {
		Name string `json:"name"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private marble name must be passed in transient map.")
	}

	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["marble_delete"]; !ok {
		return shim.Error("marble_delete must be a key in the transient map")
	}

	if len(transMap["marble_delete"]) == 0 {
		return shim.Error("marble_delete value in the transient map must be a non-empty JSON string")
	}

	var marbleDeleteInput marbleDeleteTransientInput
	err = json.Unmarshal(transMap["marble_delete"], &marbleDeleteInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(transMap["marble_delete"]))
	}

	if len(marbleDeleteInput.Name) == 0 {
		return shim.Error("name field must be a non-empty string")
	}

	// to maintain the color~name index, we need to read the marble first and get its color
	valAsbytes, err := stub.GetPrivateData("collectionMarbles", marbleDeleteInput.Name) //get the marble from chaincode state
	if err != nil {
		return shim.Error("Failed to get state for " + marbleDeleteInput.Name)
	} else if valAsbytes == nil {
		return shim.Error("Marble does not exist: " + marbleDeleteInput.Name)
	}

	var marbleToDelete marble
	err = json.Unmarshal([]byte(valAsbytes), &marbleToDelete)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(valAsbytes))
	}

	// delete the marble from state
	err = stub.DelPrivateData("collectionMarbles", marbleDeleteInput.Name)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// Also delete the marble from the color~name index
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marbleToDelete.Color, marbleToDelete.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.DelPrivateData("collectionMarbles", colorNameIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// Finally, delete private details of marble
	err = stub.DelPrivateData("collectionMarblePrivateDetails", marbleDeleteInput.Name)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// ===========================================================
// transfer a marble by setting a new owner name on the marble
// ===========================================================
func (t *SimpleChaincode) transferMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("- start transfer marble")

	type marbleTransferTransientInput struct {
		Name  string `json:"name"`
		Owner string `json:"owner"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private marble data must be passed in transient map.")
	}

	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["marble_owner"]; !ok {
		return shim.Error("marble_owner must be a key in the transient map")
	}

	if len(transMap["marble_owner"]) == 0 {
		return shim.Error("marble_owner value in the transient map must be a non-empty JSON string")
	}

	var marbleTransferInput marbleTransferTransientInput
	err = json.Unmarshal(transMap["marble_owner"], &marbleTransferInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(transMap["marble_owner"]))
	}

	if len(marbleTransferInput.Name) == 0 {
		return shim.Error("name field must be a non-empty string")
	}
	if len(marbleTransferInput.Owner) == 0 {
		return shim.Error("owner field must be a non-empty string")
	}

	marbleAsBytes, err := stub.GetPrivateData("collectionMarbles", marbleTransferInput.Name)
	if err != nil {
		return shim.Error("Failed to get marble:" + err.Error())
	} else if marbleAsBytes == nil {
		return shim.Error("Marble does not exist: " + marbleTransferInput.Name)
	}

	marbleToTransfer := marble{}
	err = json.Unmarshal(marbleAsBytes, &marbleToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	marbleToTransfer.Owner = marbleTransferInput.Owner //change the owner

	marbleJSONasBytes, _ := json.Marshal(marbleToTransfer)
	err = stub.PutPrivateData("collectionMarbles", marbleToTransfer.Name, marbleJSONasBytes) //rewrite the marble
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferMarble (success)")
	return shim.Success(nil)
}

// ===========================================================================================
// getMarblesByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) getMarblesByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetPrivateDataByRange("collectionMarbles", startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getMarblesByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryMarblesByOwner queries for marbles based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryMarblesByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	owner := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"marble\",\"owner\":\"%s\"}}", owner)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryMarbles uses a query string to perform a query for marbles.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryMarblesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryMarbles(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetPrivateDataQueryResult("collectionMarbles", queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func find(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return -1
}
