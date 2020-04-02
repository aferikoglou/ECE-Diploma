#!/bin/bash

################################################################
# A simple script for Kubernetes custom-scheduler development. #
################################################################

CUSTOM_SCHEDULER_DEP_YAML_PATH='/root/my-yaml-files/custom-scheduler/custom-scheduler-dep.yaml'

echo " "
echo "DELETING custom-scheduler DEPLOYMENT..."
echo " "
kubectl delete -f ${CUSTOM_SCHEDULER_DEP_YAML_PATH}

echo " "
echo "CREATING IMAGE FOR custom-scheduler..."
echo " "
make docker-image

echo " "
echo "PUSHING IMAGE OF custom-scheduler TO DOCKERHUB..."
echo " "
make docker-push

echo " "
echo "DEPLOYING custom-scheduler DEPLOYMENT..."
echo " "
kubectl apply -f ${CUSTOM_SCHEDULER_DEP_YAML_PATH}
