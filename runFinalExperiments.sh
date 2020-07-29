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
OVERPROV_PERCENTAGES='' # '0 50 100 150 200 250'
for OVERPROV_PERC in ${OVERPROV_PERCENTAGES}
do
        echo ''
	echo 'MIN = '${MIN}' / MAX = '${MAX}' / OVERPROVISIONING PERCENTAGE = '${OVERPROV_PERC}
        echo ''
	${EXEC_EXPERIMENT_PATH} ${MIN} ${MAX} ${OVERPROV_PERC}
done

MIN=10
MAX=20
OVERPROV_PERCENTAGES='' # '0 50 100 150 200 250'
for OVERPROV_PERC in ${OVERPROV_PERCENTAGES}
do
	echo ''
 	echo 'MIN = '${MIN}' / MAX = '${MAX}' / OVERPROVISIONING PERCENTAGE = '${OVERPROV_PERC}
 	echo ''
	${EXEC_EXPERIMENT_PATH} ${MIN} ${MAX} ${OVERPROV_PERC}
done

##### Homogeneous Workload : resnet #####

echo 'HOMOGENEOUS WORKLOAD (resnet)' 
cp ${CONFIG_FILES_PATH}config_homogeneous_workload_resnet.json ${CONFIG_FILES_PATH}config.json
 
MIN=20
MAX=40
OVERPROV_PERCENTAGES='' # '0 50 100 150 200 250'
for OVERPROV_PERC in ${OVERPROV_PERCENTAGES}
do
        echo ''
	echo 'MIN = '${MIN}' / MAX = '${MAX}' / OVERPROVISIONING PERCENTAGE = '${OVERPROV_PERC}
        echo ''
	${EXEC_EXPERIMENT_PATH} ${MIN} ${MAX} ${OVERPROV_PERC}
done    

MIN=10
MAX=20
OVERPROV_PERCENTAGES='' # '0 50 100 150 200 250'
for OVERPROV_PERC in ${OVERPROV_PERCENTAGES}
do
        echo ''
	echo 'MIN = '${MIN}' / MAX = '${MAX}' / OVERPROVISIONING PERCENTAGE = '${OVERPROV_PERC}
        echo ''
	${EXEC_EXPERIMENT_PATH} ${MIN} ${MAX} ${OVERPROV_PERC}
done

##### Heterogeneous Workload #####

echo ''
echo 'HETEROGENEOUS WORKLOAD'
echo ''
cp ${CONFIG_FILES_PATH}config_heterogeneous_workload.json ${CONFIG_FILES_PATH}config.json
     
MIN=20
MAX=40
OVERPROV_PERCENTAGES='' # '0 50 100 150 200 250'
for OVERPROV_PERC in ${OVERPROV_PERCENTAGES}
do
        echo ''
	echo 'MIN = '${MIN}' / MAX = '${MAX}' / OVERPROVISIONING PERCENTAGE = '${OVERPROV_PERC}
        echo ''
	${EXEC_EXPERIMENT_PATH} ${MIN} ${MAX} ${OVERPROV_PERC}
done        
     
MIN=10
MAX=20
OVERPROV_PERCENTAGES='' # '0 50 100 150 200 250'
for OVERPROV_PERC in ${OVERPROV_PERCENTAGES}
do
        echo ''
	echo 'MIN = '${MIN}' / MAX = '${MAX}' / OVERPROVISIONING PERCENTAGE = '${OVERPROV_PERC}
        echo ''
	${EXEC_EXPERIMENT_PATH} ${MIN} ${MAX} ${OVERPROV_PERC}
done


