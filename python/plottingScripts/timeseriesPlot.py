import os
import pandas
import argparse
import matplotlib

matplotlib.use('Agg')

from matplotlib import pyplot as plt

##### Parse command line arguments #####
parser = argparse.ArgumentParser(description='A script that plots the exported GPU metrics.')
parser.add_argument('--WORKLOAD_DIR', type=str, required=True, help='The path to the workload directory.')
# parser.add_argument('--WORKLOAD_NAME', type=str, required=True, help='The workload name.')
parser.add_argument('--METRIC', type=str, required=True, help='The GPU metric that is going to be plotted.')
 
args = parser.parse_args()
 
WORKLOAD_DIR  = args.WORKLOAD_DIR
WORKLOAD_NAME = WORKLOAD_DIR.split('/')[4] # args.WORKLOAD_NAME
METRIC        = args.METRIC

##### Main script #####
os.chdir(WORKLOAD_DIR + '/' + METRIC)

##### Create metric plot #####
series = pandas.read_csv('combined.csv', header=0, index_col=0, parse_dates=True, squeeze=True)

series.plot()

plt.title(METRIC + ' plot')
plt.xlabel('Time (s)')
	
if METRIC == 'dcgm_gpu_utilization':	
	plt.ylabel(METRIC + ' (%)')
elif METRIC == 'dcgm_power_usage':
	plt.ylabel(METRIC + ' (W)')
elif METRIC == 'dcgm_fb_used' or METRIC == 'dcgm_fb_free':
	plt.ylabel(METRIC + ' (MiB)')
elif METRIC == 'dcgm_memory_clock' or METRIC == 'dcgm_sm_clock':
	plt.ylabel(METRIC + ' (MHz)')

plt.savefig(WORKLOAD_NAME + '-' + METRIC + '.png')
