#!/bin/bash

###################################################################################
# This script is created in order to simply switch between the Alibaba's gpushare #
# scheduler extender mechanism and the default.                                   #
###################################################################################

MODE=${1}

if [ "${MODE}" == "default" ]; 
then
	echo "Setting up default k8s GPU mechanism..."

	cd /etc/kubernetes/manifests

	kubectl delete -f device-plugin-rbac.yaml
	kubectl delete -f device-plugin-ds.yaml

	# Wait for 10 seconds
	sleep 10

	kubectl create -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/1.0.0-beta4/nvidia-device-plugin.yml

elif [ "${MODE}" == "gpu-share" ];
then
	echo "Setting up gpushare scheduler extender..."

	kubectl delete ds -n kube-system nvidia-device-plugin-daemonset

	# Wait for 10 seconds
	sleep 10

	cd /etc/kubernetes/manifests

	kubectl create -f device-plugin-rbac.yaml
	kubectl create -f device-plugin-ds.yaml
else
	echo "Not supported. There are two possible modes a) default and b) gpu-share."
fi
