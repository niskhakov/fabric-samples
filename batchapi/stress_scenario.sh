#!/bin/bash


s() {
    sleep $1
}

put() {
    ./cli_stressPut.sh "$@" | tee -a stress.log
}

get() {
    ./cli_stressGet.sh "$@" | tee -a stress.log
}

del() {
    ./cli_stressDel.sh "$@" | tee -a stress.log
}

START=100
END=10000
STEP=300

KEYLEN=20
SEED=1

for i in $(seq $START $STEP $END); do
    let SEED++
    put -k $KEYLEN -s $SEED $i
    s 4
done



# put 100;
# del 100;

# put -n 100; 
# del -n 100;

# put 200;
# put -n 200;

# put 500;
# put -n 500;

# put 1000;
# put -n 1000;

# put 2000;
# put -n 2000;

# put 5000;
# del 5000;
# put -n 5000;
# del -n 5000;

# put 7000;
# put -n 7000;

# put 10000;
# del 10000;
# put -n 10000;
# del -n 10000;
 
# s 3
# put 10000;

# get 100;
# get -n 100;

# get 200;
# get -n 200;

# get 500;
# get -n 500;

# get 1000;
# get -n 1000;

# get 5000;
# get -n 5000;

# get 10000;
# get -n 10000;