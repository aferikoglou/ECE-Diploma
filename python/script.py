##########################################################################################
# This script sends workloads to my Kubernetes cluster and uses my Go monitoring program #
# in order to get GPU metrics.								 #
##########################################################################################

import os, os.path
import time
import threading

from random import seed
from random import randint

from Job import *

GO_MONITORING_PROGRAM = '/root/go/bin/prometheusGetter'
PYTHON_PLOTTING_PROGRAM = '/root/python/timeseries_plot.py'
PATH_TO_YAML_DIR = '/root/my-yaml-files'

WORKLOAD_DIR = 'w_15_100overprov'
METRICS = ['dcgm_gpu_utilization', 'dcgm_fb_used', 'dcgm_power_usage']


##### Helping Functions #####

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else: raise

def workloadGenerator(JobList):
	seed(2020)
	
	d = -1
	for index, job in enumerate(JobList, start=1):
		d = randint(30, 60)
		
		job.apply_job()

		time.sleep(d)

def exportMetrics(startTimestamp, endTimestamp, workloadOutputDirPath, workloadStartTimestamp):
	for metric in METRICS:
		outputFilePath = workloadOutputDirPath + '/' + metric + '/' + str(startTimestamp) + '_' + str(endTimestamp) + '.csv'
		command = GO_MONITORING_PROGRAM + ' -metric=' + metric + ' -range_query=true -start_timestamp=' + str(startTimestamp) + ' -end_timestamp=' + str(endTimestamp) + ' -export_to_file=true -output_file_path=' + outputFilePath + ' -workload_start_timestamp=' + str(workloadStartTimestamp)
		os.system(command)

def plotMetrics(workloadOutputDirPath):
	for metric in METRICS:
		command = 'python ' + PYTHON_PLOTTING_PROGRAM + ' ' + workloadOutputDirPath + ' ' + metric
		os.system(command)

##### Main script #####

##### Create the Job list #####
JobList = []

for f in os.listdir(PATH_TO_YAML_DIR + '/' + WORKLOAD_DIR):
	if f.endswith('.yaml'):
		path = PATH_TO_YAML_DIR + '/' + WORKLOAD_DIR + '/' + f
		newJob = Job(path)
		JobList.append(newJob)
	else:
		print('Not a .yaml file.')

##### Generate the workload #####
workGenThread = threading.Thread(target=workloadGenerator, args=(JobList,))

workGenThread.start()

##### Get GPU metrics until all jobs are finished #####
workloadStartTimestamp = int(time.time())

# Create output file path
workloadOutputFilePath = '/root/monitoring-files/TEST_0/' + WORKLOAD_DIR + '_' + str(workloadStartTimestamp) + '.txt'
# Create output directories
workloadOutputDirPath = '/root/monitoring-files/TEST_0/' + WORKLOAD_DIR + '_' + str(workloadStartTimestamp) 
mkdir_p(workloadOutputDirPath)
for metric in METRICS:
	mkdir_p(workloadOutputDirPath + '/' + metric)

startTimestamp = workloadStartTimestamp
currentTimestamp = workloadStartTimestamp

while len(JobList) > 0:
	if currentTimestamp - startTimestamp > 400:
		# Execute Go monitoring program
		exportMetrics(startTimestamp, currentTimestamp, workloadOutputDirPath, workloadStartTimestamp)

		startTimestamp = currentTimestamp + 1
	
	time.sleep(5)

	# Remove Finished Jobs from JobList
	newJobList = []
	for job in JobList:
		if job.is_applied() and job.is_completed():
			# Write Job results to file
			with open(workloadOutputFilePath, "a") as f:
				f.write('Job Name       = ' + job.get_name() + '\n')
				f.write('Execution Time = ' + str(job.get_execution_time()) + '\n')
				f.write('\n')
			# Delete Job
			job.delete_job()
		else:
			newJobList.append(job)

	JobList = newJobList

	currentTimestamp = int(time.time())

endTimestamp = int(time.time())

exportMetrics(startTimestamp, endTimestamp, workloadOutputDirPath, workloadStartTimestamp)

##### Plot output data #####
plotMetrics(workloadOutputDirPath)

