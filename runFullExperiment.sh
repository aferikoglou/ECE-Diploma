#!/bin/bash

EXEC_EXPERIMENT_PATH='/root/scripts/execExperiment.sh'

CONFIG_FILES_PATH='/root/python/YAMLGenerator/config-files/'

##### Homogeneous Workload : ssd-mobilenet #####

echo ''
echo 'HOMOGENEOUS WORKLOAD (ssd-mobilenet)'
echo ''
cp ${CONFIG_FILES_PATH}config_homogeneous_workload_ssdmobilenet.json ${CONFIG_FILES_PATH}config.json

MIN=20
MAX=40
OVERPROV_PERCENTAGES='0 50 100 150 200 250'
for OVERPROV_PERC in ${OVERPROV_PERCENTAGES}
do
        echo ''
	echo 'MIN = '${MIN}' / MAX = '${MAX}' / OVERPROVISIONING PERCENTAGE = '${OVERPROV_PERC}
        echo ''
	${EXEC_EXPERIMENT_PATH} ${MIN} ${MAX} ${OVERPROV_PERC}
done

