#!/bin/bash

docker rm -f $(docker ps -aq)
if [ $? -ne 0 ]; then
	echo "delete containers failed..."
else
	echo "successful"	
fi

rm -rf channel-artifacts/* crypto-config chainData

DOCKER_IMAGE_IDS=$(docker images | awk '($1 ~ /dev-peer.*.*.*/) {print $3}')
docker rmi -f $DOCKER_IMAGE_IDS

#echo "clear containers and images successful"
