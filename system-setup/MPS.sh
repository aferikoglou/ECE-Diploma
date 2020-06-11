#!/bin/bash

##### Script for enabling and disabling MPS #####

MODE=${1}

GPU_ID_STR="0"
GPU_ID=0

##### Define environmental variables #####
export CUDA_VISIBLE_DEVICES=${GPU_ID_STR}
# export CUDA_MPS_PIPE_DIRECTORY='/tmp/nvidia-mps'
# export CUDA_MPS_LOG_DIRECTORY='/var/log/nvidia-mps'
# export CUDA_DEVICE_MAX_CONNECTIONS=
# export CUDA_MPS_ACTIVE_THREAD_PERCENTAGE=

if [ "${MODE}" == "ENABLE" ];
then
	echo "ENABLING MPS"

	nvidia-smi -i ${GPU_ID} -c EXCLUSIVE_PROCESS

	nvidia-cuda-mps-control -d
elif [ "${MODE}" == "DISABLE" ];
then
	echo "DISABLING MPS"

	echo quit | nvidia-cuda-mps-control

	nvidia-smi -i ${GPU_ID} -c DEFAULT
elif [ "${MODE}" == "TEST" ];
then
	echo "TESTING"

	nvidia-smi -q -d compute

	ps -ef | grep mps
else
	echo "Not supported. There are three possible modes:"
	echo "a) ENABLE"
	echo "b) DISABLE"
	echo "c) TEST"
fi
