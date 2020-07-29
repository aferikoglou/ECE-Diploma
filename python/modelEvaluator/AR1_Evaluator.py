# Based on : https://www.dataquest.io/blog/understanding-regression-error-metrics/

import os
import re
import math
import argparse

from glob import glob

import matplotlib

matplotlib.use('Agg')
 
from matplotlib import pyplot as plt

##### Helping functions #####

def mean_absolute_error(predictedValList, realValList):
	summary = 0	
	n = len(predictedValList)
	for idx in range(n):
		residual = realValList[idx] - predictedValList[idx]
		summary += abs(residual)
	MAE = float(summary) / float(n)
	return MAE

def mean_square_error(predictedValList, realValList):
	summary = 0	
	n = len(predictedValList)
	for idx in range(n):
		residual = realValList[idx] - predictedValList[idx]
		summary += residual * residual
	MSE = float(summary) / float(n)
	return MSE

def mean_absolute_percentage_error(predictedValList, realValList):
	summary = 0.0	
	n = len(predictedValList)
	for idx in range(n):
		residual = realValList[idx] - predictedValList[idx]
		summary += abs(residual / realValList[idx])
	MAPE = (float(summary) / float(n)) * 100
	return MAPE

def mean_percentage_error(predictedValList, realValList):
	summary = 0.0	
	n = len(predictedValList)
	for idx in range(n):
		residual = realValList[idx] - predictedValList[idx]
		summary += residual / realValList[idx]
	MPE = (float(summary) / float(n)) * 100
	return MPE

def create_line_plot(SCHEDULER_LOGS_DIR, xValList, yValList, yValLabel):
	plt.plot(xValList, yValList, marker='.', markerfacecolor='green', markersize=6, color='green', linewidth=2)

	if yValLabel == 'predVal':
		plt.title('Free GPU Memory Pred VS Unix Timestamp', fontweight='bold')
		plt.ylabel('Free GPU Memory Prediction', fontweight='bold')
	elif yValLabel == 'realVal':
		plt.title('Real Free GPU Memory VS Unix Timestamp', fontweight='bold')
		plt.ylabel('Free GPU Memory Prediction', fontweight='bold')

	plt.xlabel('Unix Timestamp', fontweight='bold')

	plt.savefig(SCHEDULER_LOGS_DIR + '/' + yValLabel + '-linePlot.png')

	plt.clf()

def create_predReal_line_plot(SCHEDULER_LOGS_DIR, predTimestampValList, predValList, realValList):
        plt.plot(predTimestampValList, predValList, label='Pred Values', marker='.', markerfacecolor='green', markersize=6, color='green', linewidth=1)
	plt.plot(predTimestampValList, realValList, label='Real Values', marker='.', markerfacecolor='red', markersize=6, color='red', linewidth=1)
 
	plt.title('Real & Predicted Value VS Unix Timestamp', fontweight='bold')

	plt.xlabel('Unix Timestamp', fontweight='bold')
	plt.ylabel('Free GPU Memory', fontweight='bold')

	plt.legend(loc='best')

	plt.savefig(SCHEDULER_LOGS_DIR + '/realPredVal-linePlot.png')

	plt.clf()

def pred_real_scatter_plot(SCHEDULER_LOGS_DIR, STP, ORD, THR, predictedValList, realValList):
	maxVal = max(max(predictedValList), max(realValList))
	minVal = min(min(predictedValList), min(realValList))
	
	axisRangeMin = minVal - 1000
	axisRangeMax = maxVal + 1000

	line = range(axisRangeMin, axisRangeMax)

	fig, ax = plt.subplots(figsize=(7,7))

	# Create the scatter plot
	ax.scatter(predictedValList, realValList, c='green')

	# Create diagonal line
	ax.plot(line, line, 'r--', label='predVal = realVal')

	# Add labels, legend and make it nicer
	ax.set_xlabel('Predicted Values', fontweight='bold')
	ax.set_ylabel('Real Values', fontweight='bold')
	ax.set_title('Predicted vs Real Values Scatter Plot (STP=' + str(STP) + '/ORD=' + str(ORD) + '/THR=' + str(THR) + ')', fontweight='bold')
	ax.set_xlim(axisRangeMin, axisRangeMax)
	ax.set_ylim(axisRangeMin, axisRangeMax)
	ax.legend()

	plt.tight_layout()
	plt.savefig(SCHEDULER_LOGS_DIR + '/STP-' + str(STP) + '-ORD-' + str(ORD) + '-THR-' + str(THR) + '-scatterPlot.png')

	plt.clf()

##### Main script #####

##### Parse command line arguments #####
parser = argparse.ArgumentParser(description='A script that creates error metrics for the LR model that is used in the Peak Prediction module of my custom scheduler.')
parser.add_argument('--SCHEDULER_LOGS_DIR', type=str, required=True, help='The path to the scheduler log directory.')
parser.add_argument('--SCHEDULER_LOGS_NAME', type=str, required=True, help='The name of the scheduler log file.')

args = parser.parse_args()

SCHEDULER_LOGS_DIR = args.SCHEDULER_LOGS_DIR
SCHEDULER_LOGS_NAME = args.SCHEDULER_LOGS_NAME

autocorrelationValList = []
predictedValList = []
predictedValTimestampList = []
realValList = []
realValTimestampList = []

steps = None
autocorrelationOrder = None
autocorrelationThr = None

strPatternPred = 'PredictedGPUfreememory'
strPatternReal = 'RealGPUfreememory'

with open(SCHEDULER_LOGS_DIR + '/' + SCHEDULER_LOGS_NAME) as fp:
	line = fp.readline()
	while line:
		l = re.sub(r"\s+", "", line)
		if strPatternPred in l:
			s = len(predictedValList)
			predictedValTimestamp = int(l.split('/')[0].split('=')[1])
			predictedVal = int(l.split('/')[1].split('=')[1].split('MiB')[0])
			if steps == None:
				steps = int(l.split('/')[2].split('=')[1].split('sec')[0])
			if autocorrelationOrder == None:
				autocorrelationOrder = int(l.split('/')[3].split('=')[1])
			if autocorrelationThr == None:
				autocorrelationThr = float(l.split('/')[5].split('=')[1])
			autocorrelationVal = float(l.split('/')[4].split('=')[1])
			predictedValTimestampList.insert(s, predictedValTimestamp)
			predictedValList.insert(s, predictedVal)
			autocorrelationValList.insert(s, autocorrelationVal)
		elif strPatternReal in l:
			s = len(realValList)
			realValTimestamp = int(l.split('/')[0].split('=')[1])
			realVal = int(l.split('/')[1].split('=')[1].split('MiB')[0])
			if realValTimestamp - predictedValTimestampList[s] == 10:
				realValTimestampList.insert(s, realValTimestamp)
				realValList.insert(s, realVal)
			else:
				del predictedValTimestampList[s]
				del predictedValList[s]
				del autocorrelationValList[s]
		line = fp.readline()

if len(predictedValList) != 0:
	fp = open(SCHEDULER_LOGS_DIR + "/model-evaluation.txt","w+")

	fp.write('TOTAL PRED               = ' + str(len(predictedValList)) + '\n\n')
	# print('TOTAL PRED = ' + str(len(predictedValList)))

	fp.write('STEPS                    = ')
	fp.write(str(steps))
	fp.write('\n\n')
	fp.write('AUTOCORRELATION ORDER    = ')
	fp.write(str(autocorrelationOrder))
	fp.write('\n\n')
	fp.write('AUTOCORRELATION THR      = ')
	fp.write(str(autocorrelationThr))
	fp.write('\n\n')
	
	# fp.write('PRED TIMEST LIST         = ')
	# fp.write(str(predictedValTimestampList))
	# fp.write('\n\n')
	# fp.write('PRED VAL LIST            = ')
	# fp.write(str(predictedValList))
	# fp.write('\n\n')
	# fp.write('AUTOCORRELATION VAL LIST = ')
	# fp.write(str(autocorrelationValList))
	# fp.write('\n\n')
	# fp.write('REAL TIMEST LIST         = ')
	# fp.write(str(realValTimestampList))
	# fp.write('\n\n')
	# fp.write('REAL VAL LIST            = ')
	# fp.write(str(realValList))
	# fp.write('\n\n')

	fp.write('PRED TIMEST')
	fp.write('\t')
	fp.write('PRED VAL')
	fp.write('\t')
	fp.write('REAL TIMEST')
	fp.write('\t')
	fp.write('REAL VAL')
	fp.write('\t')
	fp.write('AUTOCOR VAL')
	fp.write('\n\n')

	for idx in range(len(predictedValList)):
		fp.write(str(predictedValTimestampList[idx]))
		fp.write('\t')
		fp.write(str(predictedValList[idx]))
		fp.write('\t')
		fp.write(str(realValTimestampList[idx]))
		fp.write('\t')
		fp.write(str(realValList[idx]))
		fp.write('\t')
		fp.write(str(autocorrelationValList[idx]))
		fp.write('\n\n')

	MAE = mean_absolute_error(predictedValList, realValList)
	fp.write('MAE              = ' + str(MAE) + '\n\n')
	# print('MAE  = ' + str(MAE))

	MSE = mean_square_error(predictedValList, realValList)
	fp.write('MSE              = ' + str(MSE) + '\n\n')
	# print('MSE  = ' + str(MSE))

	RMSE = math.sqrt(MSE)
	fp.write('RMSE             = ' + str(RMSE) + '\n\n')
	# print('RMSE = ' + str(RMSE))

	MAPE = mean_absolute_percentage_error(predictedValList, realValList)
	fp.write('MAPE             = ' + str(MAPE) + ' %\n\n')
	# print('MAPE = ' + str(MAPE) + ' %')

	MPE = mean_percentage_error(predictedValList, realValList)
	fp.write('MPE              = ' + str(MPE) + ' %\n\n')
	# print('MPE  = ' + str(MPE) + ' %')

	pred_real_scatter_plot(SCHEDULER_LOGS_DIR, steps, autocorrelationOrder, autocorrelationThr, predictedValList, realValList)

	# create_line_plot(SCHEDULER_LOGS_DIR, predictedValTimestampList, predictedValList, 'predVal')
	# create_line_plot(SCHEDULER_LOGS_DIR, realValTimestampList, realValList, 'realVal')
	create_predReal_line_plot(SCHEDULER_LOGS_DIR, predictedValTimestampList, predictedValList, realValList)

	fp.close()

