#!/bin/bash

###################################################################################
# This script submits workoads to my Kubernetes cluster. It uses my Go monitoring #
# program to save the values of the desired GPU metrics in different  folders  in # 
# /root/monitoring-files folder. In the end, it deletes the submitted workloads.  #
###################################################################################

YAML_DIR_PATH='/root/my-yaml-files/workloads/'
GO_MONITORING_PROGRAM='/root/go/bin/PrometheusHttpRequests'
PYTHON_PLOTTING_PROGRAM='/root/python/timeseries_plot.py'

WORKLOAD_YAML_FILES='lstmep1-gpu-workload.yaml tfresnet50ss-gpu-workload.yaml matmul32768-gpu-workload.yaml onnxmobilenetss-gpu-workload.yaml tfssdmobilenetss-gpu-workload.yaml'
METRICS='dcgm_gpu_utilization dcgm_mem_copy_utilization dcgm_power_usage dcgm_gpu_temp'

# First parameter - start_timestamp
# Second parameter - end_timestamp
# Third parameter - workload_output_dir
# Fourth parameter - workload_start_timestamp
export_metrics () {
	for metric in ${METRICS};
	do
		# Execute Go program to export data from Prometheus TSDB to file.
		${GO_MONITORING_PROGRAM} -metric=${metric} -range_query=true -start_timestamp=${1} -end_timestamp=${2} -export_to_file=true -output_file_path=${3}'/'${metric}'/'${1}'_'${2}'.csv' -workload_start_timestamp=${4}
	done
}

# Parameter - workload_output_dir
plot_metrics () {
        for metric in ${METRICS};
        do
        	# Execute Python program to plot the exported data.
		python ${PYTHON_PLOTTING_PROGRAM} ${1} ${metric}
        done
}

# Main script
for workload_yaml_file in ${WORKLOAD_YAML_FILES};
do
	full_path=${YAML_DIR_PATH}${workload_yaml_file}
        kubectl apply -f ${full_path}
	
	# Split workload yaml file by . to get workload name
	IFS='.' read -ra A <<< "$workload_yaml_file"	
	workload_name="${A[0]}"
	
	start_timestamp=$(date +"%s")
	workload_start_timestamp=${start_timestamp}
	current_timestamp=${start_timestamp}
	
	workload_output_dir='/root/monitoring-files/'${workload_name}'-'${start_timestamp}

	# Create output directories
	mkdir -p ${workload_output_dir}
	for metric in ${METRICS};
	do
		mkdir -p ${workload_output_dir}'/'${metric}
	done

	until kubectl get jobs ${workload_name} -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}' | grep -q True;
	do
		if ((${current_timestamp} - ${start_timestamp} > 500))
		then
			export_metrics ${start_timestamp} ${current_timestamp} ${workload_output_dir} ${workload_start_timestamp}
			
			start_timestamp=${current_timestamp}
			start_timestamp=$((start_timestamp+1)) 
		fi

		current_timestamp=$(date +"%s")
	done

	end_timestamp=$(date +"%s")
	
	export_metrics ${start_timestamp} ${end_timestamp} ${workload_output_dir} ${workload_start_timestamp}

	kubectl delete -f ${full_path}

	plot_metrics ${workload_output_dir}
done
