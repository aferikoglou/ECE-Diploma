import os
import sys
import glob
import pandas
import matplotlib
matplotlib.use('Agg')

from matplotlib import pyplot as plt

# Parse command line arguments
WORKLOAD_OUTPUT_DIR_PATH=sys.argv[1]
METRIC=sys.argv[2]

# Main script
os.chdir(WORKLOAD_OUTPUT_DIR_PATH + '/' + METRIC)

parts = WORKLOAD_OUTPUT_DIR_PATH.split('/')
workloadName = parts[3]	

fileList=glob.glob("*.csv")

dfList=[]
for filename in sorted(fileList):
	df=pandas.read_csv(filename,header=None)
	dfList.append(df)
	
concatDf=pandas.concat(dfList,axis=0)

concatDf.to_csv("combined.csv", index=None, header=False, encoding='utf-8-sig')

series = pandas.read_csv('combined.csv', header=0, index_col=0, parse_dates=True, squeeze=True)

series.plot()

plt.title(METRIC + ' plot' + ' (' + workloadName + ')')
plt.xlabel('Time (s)')
	
if METRIC == 'dcgm_gpu_utilization' or METRIC == 'dcgm_mem_copy_utilization':	
	plt.ylabel(METRIC + ' (%)')
elif METRIC == 'dcgm_gpu_temp':
	plt.ylabel(METRIC + ' (oC)')
elif METRIC == 'dcgm_power_usage':
	plt.ylabel(METRIC + ' (W)')

plt.savefig(workloadName + '-' + METRIC + '.png')
