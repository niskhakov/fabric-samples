/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
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
	defaultSeed      = 1
	defaultKeyLength = 7
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
	case "getObjectsBatch":
		//get multiple objects via one request
		return t.getObjectsBatch(stub, args)
	case "getManyObjectsBatch":
		//get multiple randomly selected objects via one request
		return t.getManyObjectsBatch(stub, args)
	case "putObjectsBatch":
		//put multiple objects via one request
		return t.putObjectsBatch(stub, args)
	case "putManyObjectsBatch":
		// stress test putting multiple objects via one request
		return t.putManyObjectsBatch(stub, args)
	case "delManyObjectsBatch":
		// stress test deleting multiple objects via one request
		return t.delManyObjectsBatch(stub, args)
	case "putRange":
		return t.putRange(stub, args)
	case "getRange":
		return t.getRange(stub, args)
	default:
		//error
		fmt.Println("invoke did not find func: " + function)
		return shim.Error("Received unknown function invocation")
	}
}

// ============================================================
// putObjectsBatch - put objects info via one network request
// ============================================================
func (t *SimpleChaincode) putObjectsBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
// putManyObjectsBatch - stress test putting random (with seed) objects info via one network request
// ============================================================
func (t *SimpleChaincode) putManyObjectsBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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
	var keyLengthParam int
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

	// check for keyLength param
	if indx := find(args, "keylength"); indx != -1 && indx+1 < len(args) {
		keyLengthParam, err = strconv.Atoi(args[indx+1])
		if err != nil {
			keyLengthParam = defaultKeyLength
		}
	} else {
		keyLengthParam = defaultKeyLength
	}

	RandReset(seedParam)
	keys := make([]string, 0)
	kvMap := make([]shim.StateKV, 0)
	for i := 0; i < keyQty; i++ {
		k := RandString(keyLengthParam)
		keys = append(keys, k)
		v := RandString(keyLengthParam)
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
		verboseMsg = fmt.Sprintf("useBatchAPI: %t, Collection: `%s`, Seed: %d, KeyLength: %d, Keys: %s", useBatchAPI, collectionParam, seedParam, keyLengthParam, strings.Join(keys, ", "))
	}

	res := fmt.Sprintf(`PutState:{"method":"put","entries":%d,"millis":%d,"keylen":%d,"batchapi":%t,"collection":"%s","seed":%d} %s`, keyQty, duration.Milliseconds(), keyLengthParam, useBatchAPI, collectionParam, seedParam, verboseMsg)

	return shim.Success([]byte(res))
}

// ============================================================
// getObjectsBatch - get objects via one network request
// ============================================================
func (t *SimpleChaincode) getObjectsBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
// getManyObjectsBatch - get many randomly selected (with seed) objects via one network request
// ============================================================
func (t *SimpleChaincode) getManyObjectsBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	var keyLengthParam int
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

	// check for keyLength param
	if indx := find(args, "keylength"); indx != -1 && indx+1 < len(args) {
		keyLengthParam, err = strconv.Atoi(args[indx+1])
		if err != nil {
			keyLengthParam = defaultKeyLength
		}
	} else {
		keyLengthParam = defaultKeyLength
	}

	RandReset(seedParam)

	keys := make([]shim.StateKey, 0)
	for i := 0; i < keyQty; i++ {
		keys = append(keys, shim.StateKey{Collection: collectionParam, Key: RandString(keyLengthParam)})

		// Use RandString one more time to be consistent with putManyObjects, which invokes RandString 2 times
		// and get the same keys as were written in put operation
		_ = RandString(keyLengthParam)
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

	// res := fmt.Sprintf("Get state queried: getting %d entries from the ledger takes %s %s", keyQty, duration.String(), verboseMsg)
	res := fmt.Sprintf(`GetState:{"method":"get","entries":%d,"millis":%d,"keylen":%d,"batchapi":%t,"collection":"%s","seed":%d} %s`, keyQty, duration.Milliseconds(), keyLengthParam, useBatchAPI, collectionParam, seedParam, verboseMsg)

	return shim.Success([]byte(res))
}

// ============================================================
// delManyObjectsBatch - deletes many randomly selected Objects (with seed) via one network request
// ============================================================
func (t *SimpleChaincode) delManyObjectsBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	var keyLengthParam int
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

	// check for keyLength param
	if indx := find(args, "keylength"); indx != -1 && indx+1 < len(args) {
		keyLengthParam, err = strconv.Atoi(args[indx+1])
		if err != nil {
			keyLengthParam = defaultKeyLength
		}
	} else {
		keyLengthParam = defaultKeyLength
	}

	RandReset(seedParam)

	keys := make([]shim.StateKey, 0)
	for i := 0; i < keyQty; i++ {
		keys = append(keys, shim.StateKey{Collection: collectionParam, Key: RandString(keyLengthParam)})

		// Use RandString one more time to be consistent with putManyObjects, which invokes RandString 2 times
		// and get the same keys as were written in put operation
		_ = RandString(keyLengthParam)
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

	// res := fmt.Sprintf("Del state invoked: deleting %d entries from the ledger takes %s %s", keyQty, duration.String(), verboseMsg)
	res := fmt.Sprintf(`DelState:{"method":"del","entries":%d,"millis":%d,"keylen":%d,"batchapi":%t,"collection":"%s","seed":%d} %s`, keyQty, duration.Milliseconds(), keyLengthParam, useBatchAPI, collectionParam, seedParam, verboseMsg)

	return shim.Success([]byte(res))
}

// ============================================================
// putRange - put many objects using BatchAPI (there is no PutStateByRange function in fabric)
// 	This function sets state objects which later will be queried by getRange (using GetStateByRange or BatchAPI)
// ============================================================
func (t *SimpleChaincode) putRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error(fmt.Errorf("Incorrect arguments. Expecting at least one argument").Error())
	}

	keyQty, err := strconv.Atoi(args[0]) // number of keys to read - required
	if err != nil {
		return shim.Error(err.Error())
	}

	objectsNum := keyQty

	valuesToPut := make([]shim.StateKV, 0)
	for i := 0; i < objectsNum; i++ {
		valuesToPut = append(valuesToPut, shim.StateKV{
			Collection: "",
			Key:        fmt.Sprintf("OBJ%05d", i),
			Value:      []byte(fmt.Sprintf(`{"test":"object","message":"hello developer!","id":"%d"}`, i)),
		})
	}

	start := time.Now()
	err = stub.PutStateBatch(valuesToPut)
	duration := time.Since(start)

	if err != nil {
		return shim.Error(err.Error())
	}

	res := fmt.Sprintf(`PutRange:{"method":"putrange","entries":%d,"millis":%d,"batchapi":true}`, objectsNum+1, duration.Milliseconds())

	return shim.Success([]byte(res))
}

// ============================================================
// getRange - queryies many objects using GetStateByRange
// ============================================================
func (t *SimpleChaincode) getRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error(fmt.Errorf("Incorrect arguments. Expecting at least one argument").Error())
	}

	keyQty, err := strconv.Atoi(args[0]) // number of keys to read - required
	if err != nil {
		return shim.Error(err.Error())
	}

	startkey := fmt.Sprintf("OBJ%05d", 0)
	endkey := fmt.Sprintf("OBJ%05d", keyQty)

	verboseFlag := false
	useBatchAPI := true
	// check for nobatchapi param
	if indx := find(args, "nobatchapi"); indx != -1 {
		useBatchAPI = false
	}

	// check for verbose param
	if find(args, "verbose") != -1 {
		verboseFlag = true
	}

	stateKeys := make([]shim.StateKey, 0)

	startInt, _ := strconv.Atoi(startkey[3:])
	endInt, _ := strconv.Atoi(endkey[3:])
	for i := startInt; i < endInt; i++ {
		stateKeys = append(stateKeys, shim.StateKey{
			Key:        fmt.Sprintf("OBJ%05d", i),
			Collection: "",
		})
	}

	var start time.Time
	var duration time.Duration
	var resKV []shim.StateKV

	if useBatchAPI {
		start = time.Now()
		resKV, err = stub.GetStateBatch(stateKeys)
		if err != nil {
			return shim.Error(err.Error())
		}
		duration = time.Since(start)
	} else {
		resKV = make([]shim.StateKV, 0)
		start = time.Now()
		iterator, err := stub.GetStateByRange(startkey, endkey)
		if err != nil {
			return shim.Error(err.Error())
		}

		defer iterator.Close()

		for iterator.HasNext() {
			queryResp, err := iterator.Next()
			if err != nil {
				return shim.Error(err.Error())
			}

			resKV = append(resKV, shim.StateKV{
				Key:        queryResp.Key,
				Collection: queryResp.Namespace,
				Value:      queryResp.Value,
			})
		}
		duration = time.Since(start)
	}

	var verbose string
	if verboseFlag {
		var verboseMsg strings.Builder
		verboseMsg.WriteString(",verbose:{")

		for _, kv := range resKV {
			fmt.Fprintf(&verboseMsg, `"%s":"%s",`, kv.Key, kv.Value)
		}

		toStr := verboseMsg.String()
		verbose = toStr[0 : len(toStr)-1]
		verbose += "}"
	}

	res := fmt.Sprintf(`GetRange:{"method":"getrange","entries":%d,"millis":%d,"batchapi":%t%s}`, len(resKV), duration.Milliseconds(), useBatchAPI, verbose)

	return shim.Success([]byte(res))

}

func find(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return -1
}
