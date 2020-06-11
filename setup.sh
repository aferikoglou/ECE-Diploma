#!/bin/bash

MECH=${1}
MODE=${2}

DEFAULT_SCHED_SETUP_DIR='/root/system-setup/default-sched-setup/'
ALIBABA_SCHED_SETUP_DIR='/root/system-setup/alibaba-sched-setup/'
CUSTOM_SCHED_SETUP_DIR='/root/system-setup/custom-sched-setup/'
IMPROVED_CUSTOM_SCHED_SETUP_DIR='/root/system-setup/improved-custom-sched-setup/'
NODE_EXPORTER_SETUP_DIR='/root/system-setup/node-exporter-setup/'
PROMETHEUS_SETUP_DIR='/root/system-setup/prometheus-setup/'

GPU_NODE_NAME='kube-gpu-1'

if [ "${MECH}" == "DEFAULT_SCHED" ];
then
	if [ "${MODE}" == "ENABLE" ];
	then
		kubectl create -f ${DEFAULT_SCHED_SETUP_DIR}'nvidia-device-plugin.yml'
	elif [ "${MODE}" == "DISABLE" ];
	then
		kubectl delete -f ${DEFAULT_SCHED_SETUP_DIR}'nvidia-device-plugin.yml'
	else
		echo "Not supported. There are two possible modes:"
		echo "a) ENABLE"
		echo "b) DISABLE"
	fi
elif [ "${MECH}" == "ALIBABA_SCHED" ];
then
	if [ "${MODE}" == "ENABLE" ];
	then
		# Note: DEFAULT_SCHED must be DISABLED
	
		cp ${ALIBABA_SCHED_SETUP_DIR}'scheduler-policy-config.json' /etc/kubernetes/

		kubectl create -f ${ALIBABA_SCHED_SETUP_DIR}'gpushare-schd-extender.yaml'

		cp ${ALIBABA_SCHED_SETUP_DIR}'modified-kube-scheduler.yaml' /etc/kubernetes/manifests/kube-scheduler.yaml

		kubectl create -f ${ALIBABA_SCHED_SETUP_DIR}'device-plugin-rbac.yaml'

		kubectl create -f ${ALIBABA_SCHED_SETUP_DIR}'device-plugin-ds.yaml'

		kubectl label node ${GPU_NODE_NAME} gpushare=true
 
		cp ${ALIBABA_SCHED_SETUP_DIR}'kubectl-inspect-gpushare' /usr/bin/
		chmod u+x /usr/bin/kubectl-inspect-gpushare
	elif [ "${MODE}" == "DISABLE" ];
	then
		rm /usr/bin/kubectl-inspect-gpushare

		kubectl label node ${GPU_NODE_NAME} gpushare-

		kubectl delete -f ${ALIBABA_SCHED_SETUP_DIR}'device-plugin-ds.yaml'

		kubectl delete -f ${ALIBABA_SCHED_SETUP_DIR}'device-plugin-rbac.yaml'

		cp ${ALIBABA_SCHED_SETUP_DIR}'default-kube-scheduler.yaml' /etc/kubernetes/manifests/kube-scheduler.yaml

		kubectl delete -f ${ALIBABA_SCHED_SETUP_DIR}'gpushare-schd-extender.yaml'

		rm /etc/kubernetes/scheduler-policy-config.json
	else
		echo "Not supported. There are two possible modes:"
		echo "a) ENABLE"
		echo "b) DISABLE"
	fi
elif [ "${MECH}" == "CUSTOM_SCHED" ];
then
	if [ "${MODE}" == "ENABLE" ];
	then
		kubectl create -f ${CUSTOM_SCHED_SETUP_DIR}'custom-scheduler-rbac.yaml'

		kubectl create -f ${CUSTOM_SCHED_SETUP_DIR}'custom-scheduler-dep.yaml'
	elif [ "${MODE}" == "DISABLE" ];
	then
		kubectl delete -f ${CUSTOM_SCHED_SETUP_DIR}'custom-scheduler-dep.yaml'

		kubectl delete -f ${CUSTOM_SCHED_SETUP_DIR}'custom-scheduler-rbac.yaml'
	else
		echo "Not supported. There are two possible modes:"
		echo "a) ENABLE"
		echo "b) DISABLE"
	fi
elif [ "${MECH}" == "IMPROVED_CUSTOM_SCHED" ];
then
	if [ "${MODE}" == "ENABLE" ];
	then
		kubectl create -f ${IMPROVED_CUSTOM_SCHED_SETUP_DIR}'improved-custom-scheduler-rbac.yaml'

		kubectl create -f ${IMPROVED_CUSTOM_SCHED_SETUP_DIR}'improved-custom-scheduler-dep.yaml'
	elif [ "${MODE}" == "DISABLE" ];
	then
		kubectl delete -f ${IMPROVED_CUSTOM_SCHED_SETUP_DIR}'improved-custom-scheduler-dep.yaml'

		kubectl delete -f ${IMPROVED_CUSTOM_SCHED_SETUP_DIR}'improved-custom-scheduler-rbac.yaml'
	else
		echo "Not supported. There are two possible modes:"
		echo "a) ENABLE"
		echo "b) DISABLE"
	fi
elif [ "${MECH}" == "NODE_EXPORTER" ];
then
	if [ "${MODE}" == "ENABLE" ];
	then
		kubectl label nodes ${GPU_NODE_NAME} hardware-type=NVIDIAGPU

		kubectl create -f ${NODE_EXPORTER_SETUP_DIR}'gpu-node-exporter-daemonset.yaml'
	elif [ "${MODE}" == "DISABLE" ];
	then
		kubectl delete -f ${NODE_EXPORTER_SETUP_DIR}'gpu-node-exporter-daemonset.yaml'

		kubectl label nodes ${GPU_NODE_NAME} hardware-type-
	else
		echo "Not supported. There are two possible modes:"
		echo "a) ENABLE"
		echo "b) DISABLE"
	fi
elif [ "${MECH}" == "PROMETHEUS" ];
then
	if [ "${MODE}" == "ENABLE" ];
	then
		kubectl create namespace monitoring

		kubectl create -f ${PROMETHEUS_SETUP_DIR}'clusterRole.yaml'

		kubectl create -f ${PROMETHEUS_SETUP_DIR}'config-map.yaml'

		kubectl create -f ${PROMETHEUS_SETUP_DIR}'prometheus-deployment.yaml'

		kubectl create -f ${PROMETHEUS_SETUP_DIR}'prometheus-service.yaml' --namespace=monitoring
	elif [ "${MODE}" == "DISABLE" ];
	then
		kubectl delete -f ${PROMETHEUS_SETUP_DIR}'prometheus-service.yaml' --namespace=monitoring

		kubectl delete -f ${PROMETHEUS_SETUP_DIR}'prometheus-deployment.yaml'

		kubectl delete -f ${PROMETHEUS_SETUP_DIR}'config-map.yaml'

		kubectl delete -f ${PROMETHEUS_SETUP_DIR}'clusterRole.yaml'

		kubectl delete namespace monitoring
	else
		echo "Not supported. There are two possible modes:"
		echo "a) ENABLE"
		echo "b) DISABLE"
	fi
else
	echo "Not supported. There are five possible mechanisms:"
	echo "a) DEFAULT_SCHED"
	echo "b) ALIBABA_SCHED"
	echo "c) CUSTOM_SCHED"
	echo "d) IMPROVED_CUSTOM_SCHED"
	echo "e) NODE_EXPORTER"
	echo "f) PROMETHEUS"
	echo "WARNING: You cannot enable DEFAULT_SCHED and ALIBABA_SCHED at the same time"
	echo "         CUSTOM_SCHED and IMPROVED_CUSTOM_SCHED require NODE_EXPORTER and PROMETHEUS"
fi
