import os
import sys
import glob
import pandas
import matplotlib
matplotlib.use('Agg')

from matplotlib import pyplot as plt

WORKLOAD_OUTPUT_DIR_PATH=sys.argv[1]
METRIC=sys.argv[2]

def main():
	# print(WORKLOAD_OUTPUT_DIR_PATH + '/' + METRIC)

	os.chdir(WORKLOAD_OUTPUT_DIR_PATH + '/' + METRIC)

	parts = WORKLOAD_OUTPUT_DIR_PATH.split('/')
	workload_name = parts[3]	

	fileList=glob.glob("*.csv")

	# print(fileList)

	dfList=[]
	for filename in fileList:
		df=pandas.read_csv(filename,header=None)
		dfList.append(df)
	concatDf=pandas.concat(dfList,axis=0)

	concatDf.to_csv("combined.csv", index=None, header=False, encoding='utf-8-sig')

	series = pandas.read_csv('combined.csv', header=0, index_col=0, parse_dates=True, squeeze=True)

	series.plot()

	plt.title(METRIC + ' plot' + ' (' + workload_name + ')')
	plt.xlabel('Timestamp')
	
	if METRIC == 'dcgm_gpu_utilization':	
		plt.ylabel(METRIC + ' (%)')
	elif METRIC == 'dcgm_gpu_temp':
		plt.ylabel(METRIC + ' (oC)')

	# print(series.head())

	plt.savefig(workload_name + '-' + METRIC + '.png')


if __name__ == "__main__":
	main()
