#!/bin/bash

echo load CREDIT data

var_acc=0
genAcc(){
    var_acc=$(($RANDOM%($max-$min+1)+$min))
}

var_amount=0
genAmount(){
    var_amount=$(($RANDOM%($max_amount-$min_amount+1)+$min_amount))
}

# --------------------Load n per 1-------------------------
domain=http://localhost:5001/add

min=1
max=10

max_amount=500
min_amount=300

for (( x=0; x<=10; x++ ))
do
    genAcc
    genAmount
    echo curl -X POST $domain -H 'Content-Type: application/json' -d '{"account_id": "ACC-'$var_acc'","type_charge": "CREDIT","amount":'$var_amount',"tenant_id": "TENANT-1"}'
    curl -X POST $domain -H 'Content-Type: application/json' -d '{"account_id": "ACC-'$var_acc'","type_charge": "CREDIT","amount":'$var_amount',"tenant_id": "TENANT-1"}'
done

