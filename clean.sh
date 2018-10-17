#!/bin/bash

docker rm -f $(docker ps -aq)
if [ $? -ne 0 ]; then
	echo "delete containers failed..."
	exit 1
else
	echo "successful"	
fi

#DOCKER_IMAGE_IDS=$(docker images | awk '($1 ~ /dev-peer.*.byfn.*/) {print $3}')
#docker rmi -f $DOCKER_IMAGE_IDS

#echo "clear containers and images successful"
