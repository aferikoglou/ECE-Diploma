#!/bin/bash

SETUP_SCRIPT_PATH='/root/setup.sh'

SCHEDULER='CUSTOM_SCHED'

echo " "
echo "DISABLE custom-scheduler..."
echo " "
${SETUP_SCRIPT_PATH} ${SCHEDULER} DISABLE

echo " "
echo "CREATING IMAGE FOR custom-scheduler..."
echo " "
make docker-image TAG='v0.1'

echo " "
echo "PUSHING IMAGE OF custom-scheduler TO DOCKERHUB..."
echo " "
make docker-push TAG='v0.1'

echo " "
echo "ENABLE custom-scheduler..."
echo " "
${SETUP_SCRIPT_PATH} ${SCHEDULER} ENABLE	

