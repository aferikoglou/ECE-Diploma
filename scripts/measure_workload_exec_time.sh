#!/bin/bash

################################################################################
# This script is created to estimate the execution time of the workloads that  #
# are going to be submitted in my Kubernetes cluster. The results are saved in #
# /root/workload_exec_dur.dat file.                                            #
################################################################################

YAML_DIR_PATH='/root/my-yaml-files/workloads/'
WORKLOAD_YAML_FILES='tfresnet50ss-gpu-workload.yaml lstmep1-gpu-workload.yaml matmul32768-gpu-workload.yaml onnxmobilenetss-gpu-workload.yaml tfssdmobilenetss-gpu-workload.yaml'

TOTAL_ITER=5
OUTPUT_FILE_PATH='/root/workload_exec_time.dat'

# Main script
if [ ! -f "$OUTPUT_FILE_PATH" ]; then
	echo 'WORKLOAD NAME DURATION (sec)' >> ${OUTPUT_FILE_PATH}
fi

for workload_yaml_file in ${WORKLOAD_YAML_FILES};
do
	full_path=${YAML_DIR_PATH}${workload_yaml_file}
        
	IFS='.' read -ra A <<< "$workload_yaml_file"	
	workload_name="${A[0]}"

	sum=0

	for (( it = 0; it < TOTAL_ITER; it++ ))
	do  
   		kubectl apply -f ${full_path}
		
		until kubectl get jobs ${workload_name} -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}' | grep -q True;
		do
			sleep 5
		done
		
		sleep 1

		start_timestamp_s=$(kubectl get -o template jobs.batch/${workload_name} --template={{.status.startTime}})
		start_timestamp=$(date -d"$start_timestamp_s" "+%s")

		sleep 1

		end_timestamp_s=$(kubectl get -o template jobs.batch/${workload_name} --template={{.status.completionTime}})
		end_timestamp=$(date -d"$end_timestamp_s" "+%s")

		echo ${workload_name}' - Iteration '${it}' Execution Time = '$((end_timestamp - start_timestamp))' sec'

		sleep 1		

		kubectl delete -f ${full_path}

		sum=$((sum + end_timestamp - start_timestamp))
	done

	avg=$((sum / TOTAL_ITER))

	echo ${workload_name}' '${avg} >> ${OUTPUT_FILE_PATH}
done
