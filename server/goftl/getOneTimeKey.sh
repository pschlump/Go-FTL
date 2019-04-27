#!/bin/bash

DeviceId="$1"
ID_URL_CONFIG="http://localhost:3118/api/get2FactorFromDeviceId?ver=1&DeviceId=";	

wget -o t1.out -O t2.out "$ID_URL_CONFIG$DeviceId"
cat t2.out
echo ""

