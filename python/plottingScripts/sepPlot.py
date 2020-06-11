import os, os.path
import glob
import pandas
import argparse
import matplotlib

matplotlib.use('Agg')

from matplotlib import pyplot as plt

##### Plot parameters #####
FIG_X_LEN_INCH  = 7
FIG_Y_LEN_INCH  = 9
FIG_FONTSIZE    = 9
TITLE_FONTSIZE  = 8
XLABEL_FONTSIZE = 8
YLABEL_FONTSIZE = 8
H_PAD           = 0.8
W_PAD           = 1.0
PAD             = 1.0

def createSinglePlot (CSVFilename, schedMech, METRIC):
	series = pandas.read_csv(CSVFilename, header=0, index_col=0, parse_dates=True, squeeze=True)

	series.plot()

	plt.title(schedMech + ' scheduler ' + METRIC + ' plot') # , fontsize=TITLE_FONTSIZE)

	plt.xlabel('Time (s)') # , fontsize=XLABEL_FONTSIZE)
	
	if   METRIC == 'dcgm_gpu_utilization':	
		plt.ylabel(METRIC + ' (%)') # , fontsize=YLABEL_FONTSIZE)
	elif METRIC == 'dcgm_power_usage':
		plt.ylabel(METRIC + ' (W)')
	elif METRIC == 'dcgm_fb_used' or METRIC == 'dcgm_fb_free':
		plt.ylabel(METRIC + ' (MiB)')
	elif METRIC == 'dcgm_memory_clock' or METRIC == 'dcgm_sm_clock':
		plt.ylabel(METRIC + ' (MHz)')
	
	plt.tight_layout(pad=PAD, w_pad=W_PAD, h_pad=H_PAD)

##### Parse command line arguments #####
parser = argparse.ArgumentParser(description='A script that plots the exported GPU metric for each scheduler in the same plot.')
parser.add_argument('--EXPERIMENT_DIR', type=str, required=True, help='The path to the experiment directory.')
parser.add_argument('--METRIC', type=str, required=True, help='The GPU metric that is going to be plotted.')

args = parser.parse_args()

EXPERIMENT_DIR  = args.EXPERIMENT_DIR
EXPERIMENT_NAME = EXPERIMENT_DIR.split('/')[3] 
METRIC          = args.METRIC

##### Main script #####
os.chdir(EXPERIMENT_DIR + '/OVERALL/' + METRIC)

CSVList       = list(glob.glob('*.csv'))
CSVListLength = len(CSVList)

plt.figure(figsize=(FIG_X_LEN_INCH,FIG_Y_LEN_INCH)) # .suptitle(EXPERIMENT_NAME, fontsize=FIG_FONTSIZE, y=1.0)

idx = 1
for CSVFilename in CSVList:
	plt.subplot(CSVListLength, 1, idx)
	schedMech = CSVFilename.split('.')[0].split('_')[1]
	createSinglePlot(CSVFilename, schedMech, METRIC)
	idx += 1

plt.savefig(EXPERIMENT_NAME + '-SEP-' + METRIC + '.png')
