#!/bin/sh

CURL="curl -s"

# user 1
ID1=$(${CURL} -X PUT -d '{"first_name":"one","last_name":"first","nickname":"the_one","email":"the_one@gang.org","password":"passw0rd"}' http://localhost:6000/api | jq .)

# user 2
ID2=$(${CURL} -X PUT -d '{"first_name":"two","last_name":"second","nickname":"the_following","email":"again@gang.org","password":"passw0rd"}' http://localhost:6000/api | jq .)

# user 3
ID3=$(${CURL} -X PUT -d '{"first_name":"three","last_name":"third","nickname":"the_revenge","email":"last@gang.org","password":"passw0rd"}' http://localhost:6000/api | jq .)

# update user 1 country
${CURL} -X POST -d '{"country":"CH"}' "http://localhost:6000/api/${ID1}" | jq
