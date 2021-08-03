# BatchAPI Example

## First Steps

- `make / make start-network`
- `make cli` <- you will be in cli docker container
- `make cli-init` <- install and instantiate test chaincode
- `make cli-batch` <- invoke and query batch methods
- `make stress` <- run stress test scenario with putting, getting, deleting up to 10000 elements 

## More Advanced
Files: `cli_stressGet.sh`, `cli_stressPut.sh`, `cli_stressDel.sh` contain logic of invoking chaincode BatchAPI mathods with provided params, such as number of keys, random seed, verbosity, not using batch api, and using private collections. See `--help`. 

If you want to see used in chaincode keys and params, use `-v` flag to get this information.   

If you don't want to use BatchAPI, use `-n` flag, which sends `nobatchapi` param. 

If you want to specify random seed, use `-s SEED` parameter, for example, `-s 54`. For every entry random function is called 2 times: 1 - key, 2 - value(if needed, else value is discarded). You can reproduce behaviour of generating keys and values. 

If you want to see private data collection results, use `-c COLLECTION_NAME` parameter. All actions on the entries will be performed on specified collection. For example, you can use `collectionMarbles` private collection, the list of available private collections can be found in `./chaincode/collections_config.json` file, for now there are two private collections. 

## Examples
```sh
(user)$ make start-network
(user)$ make cli
(cli) $ make cli-init   # Install and instantiate chaincode
(cli) $ ./cli_stressPut.sh 1000 # Writes 1000 entries with random keys and values to the ledger
(cli) $ ./cli_stressPut.sh -s 4 -c collectionMarbles 2000 # Writes 2000 key/values to the private collection: `collectionMarbles` with rand seed
(cli) $ ./cli_stressPut.sh -n 1000 # Writes 1000 entries without using BatchAPI, for every entry PutState is called
(cli) $ ./cli_stressPut.sh -v 100 # Writes 100 entries and outputs keys
(cli) $ ./cli_stressGet.sh -v 100 # Queries ledger about 100 entries, which were written in previous request (if seed is same)
(cli) $ ./cli_stressDel.sh -n 100 # Delete entries without using BatchAPI
(cli) $ exit    # Exit the `cli` docker container
(user)$ make stop-network
```

You can see put/get/del duration times by running `stress_scenario.sh` in cli container (make cli)
```sh
(user)$ make start-network
(user)$ make cli
(cli) $ make stress # or make cli-init stress-scenario 
(cli) $ exit
(user)$ make stop-network
```