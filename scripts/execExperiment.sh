#!/bin/bash

##### Paths #####
YAML_PATH='/root/my-yaml-files/'
MONITORING_FILES_PATH='/root/monitoring-files/'
CONFIG_FILES_PATH='/root/python/YAMLGenerator/config-files/'

##### Helping Programs Paths #####
YAML_GENERATOR_PATH='/root/python/YAMLGenerator/YAMLGenerator.py'
SETUP_SCRIPT_PATH='/root/setup.sh'
START_EXPERIMENT_PATH='/root/python/startExperiment.py'
TIMESERIES_PLOT_PATH='/root/python/plottingScripts/timeseriesPlot.py'
OVERALL_PLOT_PATH='/root/python/plottingScripts/overallPlot.py'
SEP_PLOT_PATH='/root/python/plottingScripts/sepPlot.py'

##### Workload Generator Parameters #####
DELAY_SEED=2020
SUFFLE_SEED=2020

##### Get command line arguments #####

##### Workload Generator Parameters #####
MIN=${1}
MAX=${2}
##### Overprovisioning Percentage #####
OVERPROV_PERC=${3}

##### Scheduling Mechanisms #####
SCHED_MECHANISMS='DEFAULT ALIBABA CUSTSCHED'
##### Define whether GPU metrics are going to be exported #####
EXPORT_METRICS='True'
##### GPU Metrics that are going to be exported #####
METRICS='dcgm_gpu_utilization dcgm_fb_used dcgm_power_usage'

EXPERIMENT_TIMESTAMP=$(date +%s)
EXPERIMENT_NAME='EXPERIMENT_'${EXPERIMENT_TIMESTAMP}'_MIN_'${MIN}'_MAX_'${MAX}'_OVERPROV_'${OVERPROV_PERC}
for METRIC in ${METRICS}
do
	mkdir -p ${MONITORING_FILES_PATH}${EXPERIMENT_NAME}'/OVERALL/'${METRIC}
done

for MECH in ${SCHED_MECHANISMS}
do
	echo 'GPU SCHEDULING MECHANISM : '${MECH}
	
	WORKLOAD_INPUT_DIR_NAME='WORKLOAD_'${MECH}
	WORKLOAD_OUTPUT_DIR_NAME=${WORKLOAD_INPUT_DIR_NAME}'_'${EXPERIMENT_TIMESTAMP}

	##### Create Workload YAML Files #####
	rm -r ${YAML_PATH}${WORKLOAD_INPUT_DIR_NAME}

	python ${YAML_GENERATOR_PATH} --OUTPUT_DIR ${YAML_PATH}${WORKLOAD_INPUT_DIR_NAME} --CONFIG_FILE ${CONFIG_FILES_PATH}'config.json' --OVERPROV_PERC ${OVERPROV_PERC} --TEMPLATE ${MECH}

	##### Start Experiment #####	

	##### Enable scheduler #####
	if [ "${MECH}" == "DEFAULT" ]; 
	then
		${SETUP_SCRIPT_PATH} DEFAULT_SCHED ENABLE
	elif [ "${MECH}" == "ALIBABA" ];
	then
		${SETUP_SCRIPT_PATH} ALIBABA_SCHED ENABLE
	elif [ "${MECH}" == "CUSTSCHED" ];
	then
		${SETUP_SCRIPT_PATH} CUSTOM_SCHED ENABLE
	fi

	sleep 45

	python ${START_EXPERIMENT_PATH} --WORKLOAD_INPUT_DIR ${YAML_PATH}${WORKLOAD_INPUT_DIR_NAME} --WORKLOAD_OUTPUT_DIR ${MONITORING_FILES_PATH}${EXPERIMENT_NAME}'/'${WORKLOAD_OUTPUT_DIR_NAME} --MIN ${MIN} --MAX ${MAX} --DELAY_SEED ${DELAY_SEED} --SUFFLE_SEED ${SUFFLE_SEED} --EXPORT_METRICS ${EXPORT_METRICS} --METRICS ${METRICS}

	##### Disable scheduler #####
	if [ "${MECH}" == "DEFAULT" ];
	then
		${SETUP_SCRIPT_PATH} DEFAULT_SCHED DISABLE
	elif [ "${MECH}" == "ALIBABA" ];
	then
		${SETUP_SCRIPT_PATH} ALIBABA_SCHED DISABLE
	elif [ "${MECH}" == "CUSTSCHED" ];
	then
		kubectl logs $(kubectl get pods | awk '/custom-scheduler/ {print $1;exit}') > ${MONITORING_FILES_PATH}${EXPERIMENT_NAME}'/custom-scheduler.log'
		${SETUP_SCRIPT_PATH} CUSTOM_SCHED DISABLE
	fi

	sleep 45

	##### Plot Metrics #####
	for METRIC in ${METRICS}
	do
		python ${TIMESERIES_PLOT_PATH} --WORKLOAD_DIR ${MONITORING_FILES_PATH}${EXPERIMENT_NAME}'/'${WORKLOAD_OUTPUT_DIR_NAME} --METRIC ${METRIC}
		
		##### Copy combined.csv files to the OVERALL directory #####
		cp ${MONITORING_FILES_PATH}${EXPERIMENT_NAME}'/'${WORKLOAD_OUTPUT_DIR_NAME}'/'${METRIC}'/combined.csv' ${MONITORING_FILES_PATH}${EXPERIMENT_NAME}'/OVERALL/'${METRIC}'/combined_'${MECH}'.csv'
	done
done

##### Create Overall Experiment Plots #####
for METRIC in ${METRICS}
do
	python ${OVERALL_PLOT_PATH} --EXPERIMENT_DIR ${MONITORING_FILES_PATH}${EXPERIMENT_NAME} --METRIC ${METRIC}
	python ${SEP_PLOT_PATH} --EXPERIMENT_DIR ${MONITORING_FILES_PATH}${EXPERIMENT_NAME} --METRIC ${METRIC}
done
