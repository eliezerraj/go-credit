# go-credit

POC for test purposes

CRUD a account_statement data synchronoius (REST)

## Diagram

go-credit (post:add/fund) == (REST) ==> go-account (service.AddFundBalanceAccount) 

## database

See repo https://github.com/eliezerraj/go-account-migration-worker.git

## Endpoints

+ GET /header

+ GET /info

+ POST /add

        {
            "account_id": "ACC-1",
            "type_charge": "CREDIT",
            "currency": "BRL",
            "amount": 100.00,
            "tenant_id": "TENANT-200"
        }

+ GET /list/ACC-1

+ GET /listPerDate?account=ACC-1&date_start=2024-07-24

## K8 local

Add in hosts file /etc/hosts the lines below

    127.0.0.1   credit.domain.com

or

Add -host header in PostMan

## AWS

Create a public apigw