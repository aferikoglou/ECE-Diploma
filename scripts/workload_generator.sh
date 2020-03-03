#!/bin/bash

YAML_DIR_PATH='/root/my-yaml-files/w_25/'
WORKLOAD_YAML_FILES=${YAML_DIR_PATH}'*'

MODE=${1}

# Main script
for yamlFile in ${WORKLOAD_YAML_FILES};
do
	echo ${yamlFile}

	if [ "${MODE}" == "create" ];
	then
		kubectl apply -f ${yamlFile}
	elif [ "${MODE}" == "delete" ];
	then
		kubectl delete -f ${yamlFile}
	else
		echo "Not supported. There are two possible modes a) create and b) delete."
	fi
done
