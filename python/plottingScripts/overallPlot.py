import os
import glob
import pandas
import argparse
import matplotlib

matplotlib.use('Agg')

from matplotlib import pyplot as plt

##### Parse command line arguments #####
parser = argparse.ArgumentParser(description='A script that plots the overall data.')
parser.add_argument('--EXPERIMENT_DIR', type=str, required=True, help='The path to the experiment directory.')
# parser.add_argument('--EXPERIMENT_NAME', type=str, required=True, help='The experiment name.')
parser.add_argument('--METRIC', type=str, required=True, help='The GPU metric that are going to be plotted.')

args = parser.parse_args()

EXPERIMENT_DIR  = args.EXPERIMENT_DIR
EXPERIMENT_NAME = EXPERIMENT_DIR.split('/')[3] # args.EXPERIMENT_NAME
METRIC          = args.METRIC

##### Main script #####
os.chdir(EXPERIMENT_DIR + '/OVERALL/' + METRIC)

fileList=glob.glob("*.csv")

for filename in fileList:
	df = pandas.read_csv(filename, index_col = 0)

	##### Get Scheduling Mechanism #####
	sched_mech = filename.split('_')[1].split('.')[0]

	plt.plot(df, label=str(sched_mech))

plt.legend()

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

plt.savefig(EXPERIMENT_NAME + '-' + METRIC + '.png')

plt.clf()
