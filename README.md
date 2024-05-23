# go-credit

POC for test purposes

CRUD a account_statement data synchronoius (REST)

## Diagram

go-credit (post:add/fund) == (REST) ==> go-account (service.AddFundBalanceAccount) 

## database

        CREATE TABLE public.account_statement (
            id serial4 NOT NULL,
            fk_account_id int4 NULL,
            type_charge varchar(200) NULL,
            charged_at timestamptz NULL,
            currency varchar(10) NULL,
            amount float8 NULL,
            tenant_id varchar(200) NULL,
            CONSTRAINT account_statement_pkey PRIMARY KEY (id)
        );

## Endpoints

+ POST /add

        {
            "account_id": "ACC-1",
            "type_charge": "CREDIT",
            "currency": "BRL",
            "amount": 100.00,
            "tenant_id": "TENANT-200"
        }

+ GET /header

+ GET /info

+ GET /list/ACC-1

        curl svc01.domain.com/list/ACC-1 | jq

## K8 local

Add in hosts file /etc/hosts the lines below

    127.0.0.1   credit.domain.com

or

Add -host header in PostMan


## AWS

Create a public apigw