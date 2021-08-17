/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway } = require('fabric-network');
const path = require('path');
const fs = require('fs');

const stressLogsDir = "stress_logs"
const defaultKeyLen = 20
const ccpPath = path.resolve(__dirname, '..', '..', 'first-network', 'connection-org1.json');

const sleep = ms => new Promise(resolve => setTimeout(resolve, ms))

async function main() {
    try {

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = new FileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        const userExists = await wallet.exists('user1');
        if (!userExists) {
            console.log('An identity for the user "user1" does not exist in the wallet');
            console.log('Run the registerUser.js application before retrying');
            return;
        }

        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        await gateway.connect(ccpPath, { wallet, identity: 'user1', discovery: { enabled: true, asLocalhost: true } });

        // Get the network (channel) our contract is deployed to.
        const network = await gateway.getNetwork('mychannel');

        // Get the contract from the network.
        const contract = network.getContract('marblesp');

        await doTestScenarioGetRangeQueryAndGetBatchAPIWithConstantEntries(contract, 1, true)

        // for (let testId = 1; testId <=4; testId++) {
        //     await doStressTestPutAndDelWithIncreasingKeyNumber(contract, testId, 100, 300, 1000)
        // }

        // for (let testId = 1; testId <= 4; testId++) {
        //     await doStressTestPutWithIncreasingKeyNumber(contract, testId, 100, 300, 1000)
        // }

        // const repeatNum = 10
        // for (let entries = 100; entries <= 1000; entries += 300) {
        //     await doStressTestPutWithSameKeyNumberNtimes(contract, entries, repeatNum)
        // }

        // Disconnect from the gateway.
        await gateway.disconnect();

    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

async function doStressTestPutWithIncreasingKeyNumber(contract, testId, start, step, end, keylength = defaultKeyLen, useBatchAPI = true, collection = '') {
    const logPath = path.resolve(__dirname, "..", stressLogsDir, `stressPut${testId}.KeyLen${keylength}.log`)
    const logFile = fs.createWriteStream(logPath, {flags: 'a'})
    for (let entries = start; entries <= end; entries += step) {
        let buf = await contract.submitTransaction('putManyMarblesBatch', `${entries}`, ...processOptions(keylength, useBatchAPI, 10, collection));
        let bufStr = buf.toString()
        logFile.write(bufStr.substring(bufStr.indexOf(":") + 1, bufStr.length - 1) + "\n")
        console.log(`Transaction has been submitted, result is: ${buf.toString()}`);
    }
    logFile.end()
}

async function doStressTestPutWithSameKeyNumberNtimes(contract, entries, nTimes, keylength = defaultKeyLen, useBatchAPI = true, collection = '') {
    const logPath = path.resolve(__dirname, "..", stressLogsDir, `stressPut${entries}Entries.KeyLen${keylength}.log`)
    const logFile = fs.createWriteStream(logPath, {flags: 'a'})
    for (let i = 0; i < nTimes; i += 1) {
        let buf = await contract.submitTransaction('putManyMarblesBatch', `${entries}`, ...processOptions(keylength, useBatchAPI, 10, collection));
        let bufStr = buf.toString()
        logFile.write(bufStr.substring(bufStr.indexOf(":") + 1, bufStr.length - 1) + "\n")
        console.log(`Transaction has been submitted, result is: ${buf.toString()}`);
    }
    logFile.end()
}

async function doStressTestPutAndDelWithIncreasingKeyNumber(contract, testId, start, step, end, keylength = defaultKeyLen, seed = 10, useBatchAPI = true, collection = '') {
    const logPath = path.resolve(__dirname, "..", stressLogsDir, `stressPutAndDel${testId}.KeyLen${keylength}.log`)
    const logFile = fs.createWriteStream(logPath, {flags: 'a'})
    for (let entries = start; entries <= end; entries += step) {
        let buf = await contract.submitTransaction('putManyMarblesBatch', `${entries}`, ...processOptions(keylength, useBatchAPI, seed, collection));
        let bufStr = buf.toString()
        let delBuf = await contract.submitTransaction('delManyMarblesBatch', `${entries}`, ...processOptions(keylength, useBatchAPI, seed, collection))
        let delStr = delBuf.toString()
        logFile.write(bufStr.substring(bufStr.indexOf(":") + 1, bufStr.length - 1) + "\n")
        logFile.write(delStr.substring(delStr.indexOf(":") + 1, delStr.length - 1) + "\n")
        console.log(`Transaction has been submitted, result is: ${bufStr}`);
        console.log(`Transaction has been submitted, result is: ${delStr}`);
    }
    logFile.end()
}

async function doTestScenarioGetRangeQueryAndGetBatchAPIWithConstantEntries(contract, testId, verbose = false) {
    const logPath = path.resolve(__dirname, "..", stressLogsDir, `scenarioGetRangeVSBatchAPI${testId}.log`)
    const logFile = fs.createWriteStream(logPath, {flags: 'a'})
    const opts = []
    if (verbose) {
        opts.push("verbose")
    }

    // Write key-values - to read them later
    let buf = await contract.submitTransaction("putRange")
    console.log(`Transaction has been submitted: result is: ${buf.toString()}`)

    // First, get keys and values via standard getStateByRange
    let readBuf = await contract.evaluateTransaction("getRange", "nobatchapi", ...opts)
    console.log(`Transaction has been evaluated: result is: ${readBuf.toString()}`)
    let bufStr = readBuf.toString() 
    logFile.write(bufStr.substring(bufStr.indexOf(":") + 1) + "\n")

    // Second, get keys and values via batch api
    readBuf = await contract.evaluateTransaction("getRange", ...opts)
    console.log(`Transaction has been evaluated: result is: ${readBuf.toString()}`)
    bufStr = readBuf.toString()
    logFile.write(bufStr.substring(bufStr.indexOf(":") + 1) + "\n")
}

function processOptions(keylength = defaultKeyLen, useBatchAPI = true, seed = 1, collection = null) {
    const transactionARGs = []

    if (keylength) {
        transactionARGs.push('keylength', keylength.toString())
    }

    if (!useBatchAPI) {
        transactionARGs.push("nobatchapi")
    }

    if (seed) {
        if (typeof seed !== 'number' || isNaN(seed)) {
            throw Error(`Expects seed is the number, got ${seed}`)
        }
        transactionARGs.push('seed', seed.toString())
    }

    if (collection) {
        transactionARGs.push('collection', collection)
    }

    console.log(transactionARGs)
    return transactionARGs
}


main();
