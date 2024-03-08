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

#domain=http://localhost:5001/add
domain=https://go-api-global.architecture.caradhras.io/credit/add

TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwic2NvcGUiOlsiYWRtaW4iXSwiZXhwIjoxNzA4MDQ2MDg2fQ.s2997s5lHtxDAfOFYZCmPOxmKmkrlDCARcMnbndfR3s

min=1
max=1000

max_amount=6500
min_amount=540

for (( x=0; x<=1000; x++ ))
do
    genAcc
    genAmount
    echo    curl -X POST $domain -H 'Content-Type: application/json' -H "Authorization: $TOKEN" -d '{"account_id": "ACC-'$var_acc'","type_charge": "CREDIT","amount":'$var_amount',"tenant_id": "TENANT-1"}'
            curl -X POST $domain -H 'Content-Type: application/json' -H "Authorization: $TOKEN" -d '{"account_id": "ACC-'$var_acc'","type_charge": "CREDIT","amount":'$var_amount',"tenant_id": "TENANT-1"}'
done

