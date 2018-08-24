#!/bin/bash

if [[ -z "${_am_local_db_orgservice_dbstring}" ]]; then
  echo "error _am_local_db_xxx_dbstring not set"
  exit
fi

ORG_ADDR=":50051"
USER_ADDR=":50052"
SG_ADDR=":50053"
ADDR_ADDR=":50054"

# run org service
docker run -d -p $ORG_ADDR:50051 -e "APP_ENV=local" -e "_am_local_db_orgservice_dbstring=${_am_local_db_orgservice_dbstring}" linkai_orgservice:latest orgservice

# run user service
docker run -d -p $USER_ADDR:50051 -e "APP_ENV=local" -e "_am_local_db_userservice_dbstring=${_am_local_db_userservice_dbstring}" linkai_userservice:latest userservice

# run scangroup service
docker run -d -p $SG_ADDR:50051 -e "APP_ENV=local" -e "_am_local_db_scangroupservice_dbstring=${_am_local_db_scangroupservice_dbstring}" linkai_scangroupservice:latest scangroupservice

# run addr service
docker run -d -p $ADDR_ADDR:50051 -e "APP_ENV=local" -e "_am_local_db_addressservice_dbstring=${_am_local_db_addressservice_dbstring}" linkai_addressservice:latest addressservice
