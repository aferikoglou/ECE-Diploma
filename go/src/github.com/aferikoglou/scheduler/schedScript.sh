#!/bin/bash

SETUP_SCRIPT_PATH='/root/setup.sh'

# SCHEDULER='CUSTOM_SCHED'
SCHEDULER='IMPROVED_CUSTOM_SCHED'

if [ "${SCHEDULER}" == "IMPROVED_CUSTOM_SCHED" ];
then
	echo " "
	echo "DISABLE improved-custom-scheduler..."
	echo " "
	${SETUP_SCRIPT_PATH} ${SCHEDULER} DISABLE

	echo " "
	echo "CREATING IMAGE FOR improved-custom-scheduler..."
	echo " "
	make docker-image TAG='v0.2'

	echo " "
	echo "PUSHING IMAGE OF improved-custom-scheduler TO DOCKERHUB..."
	echo " "
	make docker-push TAG='v0.2'

	echo " "
	echo "ENABLE improved-custom-scheduler..."
	echo " "
	${SETUP_SCRIPT_PATH} ${SCHEDULER} ENABLE

elif [ "${SCHEDULER}" == "CUSTOM_SCHED" ];
then
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
fi
