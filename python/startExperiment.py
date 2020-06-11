import os, os.path
import time
import glob
import pandas
import random
import argparse
import threading

from shutil import copyfile
from Job import *

GO_MONITORING_PROGRAM = '/root/go/bin/prometheusGetter'

##### Helping Functions #####

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else: raise

def str2bool(v):
    if isinstance(v, bool):
       return v
    if v.lower() in ('yes', 'true', 't', 'y', '1'):
        return True
    elif v.lower() in ('no', 'false', 'f', 'n', '0'):
        return False
    else:
        raise argparse.ArgumentTypeError('Boolean value expected.')

def workloadGenerator(JobList, MIN, MAX, DELAY_SEED, SUFFLE_SEED):
	##### Define seed for the random number generator #####
	random.seed(DELAY_SEED)
	##### Suffle JobList #####
	random.Random(SUFFLE_SEED).shuffle(JobList)
	for index, job in enumerate(JobList, start=1):
		d = random.randint(MIN, MAX)
		job.apply_job()
		time.sleep(d)

def exportMetrics(startTimestamp, endTimestamp, workloadOutputDirPath, workloadStartTimestamp):
	for metric in METRICS:
		outputFilePath = workloadOutputDirPath + '/' + metric + '/' + str(startTimestamp) + '_' + str(endTimestamp) + '.csv'
		command = GO_MONITORING_PROGRAM + ' -metric=' + metric + ' -range_query=true -start_timestamp=' + str(startTimestamp) + ' -end_timestamp=' + str(endTimestamp) + \
 ' -export_to_file=true -output_file_path=' + outputFilePath + ' -workload_start_timestamp=' + str(workloadStartTimestamp)
		os.system(command)

##### Main script #####

##### Parse Command Line Arguments #####
parser = argparse.ArgumentParser(description='A script that executes an experiment. Sends a workload to my Kubernetes cluster and exports GPU metrics.')
parser.add_argument('--WORKLOAD_INPUT_DIR', type=str, required=True, help='The path to the input directory.')
parser.add_argument('--WORKLOAD_OUTPUT_DIR', type=str, required=True, help='The path to the output directory.')
parser.add_argument('--MIN', type=int, required=True, help='The minimum delay between the creation of two Jobs.')
parser.add_argument('--MAX', type=int, required=True, help='The maximum delay between the creation of two Jobs.')
parser.add_argument('--DELAY_SEED', type=int, required=True, help='The seed for the delay random number generator.')
parser.add_argument('--SUFFLE_SEED', type=int, required=True, help='The seed for the Job list suffling.')
parser.add_argument('--EXPORT_METRICS', type=str2bool, required=True, help='Defines whether the GPU metrics are going to be exported.')
parser.add_argument('--METRICS', nargs='+', type=str, required=True, help='The GPU metrics that are going to be exported.')

args = parser.parse_args()

WORKLOAD_INPUT_DIR  = args.WORKLOAD_INPUT_DIR
WORKLOAD_OUTPUT_DIR = args.WORKLOAD_OUTPUT_DIR
MIN                 = args.MIN
MAX                 = args.MAX
DELAY_SEED          = args.DELAY_SEED
SUFFLE_SEED         = args.SUFFLE_SEED
EXPORT_METRICS      = args.EXPORT_METRICS
METRICS             = args.METRICS

##### Create the Job list #####
JobList = []

for f in os.listdir(WORKLOAD_INPUT_DIR):
	if f.endswith('.yaml'):
		path = WORKLOAD_INPUT_DIR + '/' + f
		newJob = Job(path)
		JobList.append(newJob)
	else:
		print('Not a .yaml file.')

##### Generate the workload #####
workGenThread = threading.Thread(target=workloadGenerator, args=(JobList, MIN, MAX, DELAY_SEED, SUFFLE_SEED,))
workGenThread.daemon = True

workGenThread.start()

##### Get GPU metrics until all jobs are finished #####
workloadStartTimestamp = int(time.time())

##### Create output file path #####
workloadOutputFilePath = WORKLOAD_OUTPUT_DIR + '.txt'

startTimestamp = workloadStartTimestamp
currentTimestamp = 0

if EXPORT_METRICS:
	##### Create output directories #####
	workloadOutputDirPath = WORKLOAD_OUTPUT_DIR 
	mkdir_p(workloadOutputDirPath)
	for metric in METRICS:
		mkdir_p(workloadOutputDirPath + '/' + metric)
	
	currentTimestamp = workloadStartTimestamp

while len(JobList) > 0:
	##### Remove Finished Jobs from JobList #####
	newJobList = []
	for job in JobList:
		if job.is_applied() and job.is_completed():
			##### Write Job results to file #####
			with open(workloadOutputFilePath, "a") as f:
				f.write(job.get_Job_stats())
			##### Delete Job #####
			job.delete_job()
		else:
			newJobList.append(job)

	JobList = newJobList

        if EXPORT_METRICS:
		if currentTimestamp - startTimestamp > 450:
			##### Execute Go monitoring program #####
			exportMetrics(startTimestamp, currentTimestamp, workloadOutputDirPath, workloadStartTimestamp)

			startTimestamp = currentTimestamp + 1

		currentTimestamp = int(time.time())

	time.sleep(5)

endTimestamp = int(time.time())

d = endTimestamp - workloadStartTimestamp
 
##### Write overall duration #####
with open(workloadOutputFilePath, "a") as f:
	f.write('Overall Duration                             = ')
	f.write(str(d))
	f.write(' sec\n')

if EXPORT_METRICS:
	exportMetrics(startTimestamp, endTimestamp, workloadOutputDirPath, workloadStartTimestamp)

	##### Concatenate .csv files to combined.csv #####
	for metric in METRICS:
		os.chdir(workloadOutputDirPath + '/' + metric)

		fileList = glob.glob("*.csv")
 
		dfList = []
		for filename in sorted(fileList):
			df = pandas.read_csv(filename, header=None)
			dfList.append(df)
    
		concatDf = pandas.concat(dfList,axis=0)
 
		concatDf.to_csv("combined.csv", index=None, header=False, encoding='utf-8-sig')

