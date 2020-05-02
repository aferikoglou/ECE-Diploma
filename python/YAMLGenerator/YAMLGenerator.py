import os
import errno    
import yaml
import json
import argparse

##### Helping Functions #####

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else:
            raise

def defaultTemplate(name, container_name, image_name, command):
         t = {
           "apiVersion": "batch/v1",
           "kind": "Job",
           "metadata": {
             "name": str(name)
            },
            "spec": {
              "parallelism": 1,
              "completions": 1,
              "template": {
                "metadata": {
                  "name": str(name) + "-pod"
                 },
                "spec": {
                  "restartPolicy": "OnFailure",
                  "containers": [
                  {
                    "name": str(container_name),
                    "image": str(image_name),
                    "command": [
                      "sh",
                      "-c",
                      str(command)
                    ],
                    "resources": {
                      "limits": {
                        "nvidia.com/gpu": 1
                       }
                     }
                   }
                 ]
               }
             }
           }
         }
 
         out = yaml.dump(t)
 
         return out



def alibabaTemplate(name, container_name, image_name, command, memory_request):
	t = {
          "apiVersion": "batch/v1",
          "kind": "Job",
          "metadata": {
            "name": str(name)
           },
           "spec": {
             "parallelism": 1,
             "completions": 1,
             "template": { 
               "metadata": {
                 "name": str(name) + "-pod"
                },
               "spec": {
                 "restartPolicy": "OnFailure",
                 "containers": [
                 {
                   "name": str(container_name),
                   "image": str(image_name),
                   "command": [
                     "sh",
                     "-c",
                     str(command)
                   ],
                   "resources": {
                     "limits": {
                       "aliyun.com/gpu-mem": memory_request
                      }
                    }
                  }
                ]
              }
            }
          }
        }

	out = yaml.dump(t)

	return out

def custschedTemplate(name, container_name, image_name, command, memory_request):
         t = {
           "apiVersion": "batch/v1",
           "kind": "Job",
           "metadata": {
             "name": str(name)
            },
            "spec": {
              "parallelism": 1,
              "completions": 1,
              "template": {
                "metadata": {
                  "name": str(name) + "-pod",
                  "labels": {
                     "GPU_MEM_REQ": str(memory_request)
                   }
                 },
                "spec": {
                  "schedulerName": "custom-scheduler",
                  "restartPolicy": "OnFailure",
                  "containers": [
                  {
                    "name": str(container_name),
                    "image": str(image_name),
                    "command": [
                      "sh",
                      "-c",
                      str(command)
                    ]
                   }
                 ]
               }
             }
           }
         }
 
         out = yaml.dump(t)
 
         return out

def getOverprovisionedMemory(optimal_memory_request_GB, percentage):
	MAX_GPU_AVAIL_MEM = 32

	overprovisioned_memory_request_GB = int(optimal_memory_request_GB + (percentage / 100) * optimal_memory_request_GB)

	return (overprovisioned_memory_request_GB if overprovisioned_memory_request_GB < MAX_GPU_AVAIL_MEM else MAX_GPU_AVAIL_MEM)

##### Main Script #####

##### Parse Command Line Arguments #####
parser = argparse.ArgumentParser(description='A script that creates the YAML files of a workload based on a JSON configuration file.')
parser.add_argument('--OUTPUT_DIR', type=str, required=True, help='The path to the output directory.')
parser.add_argument('--CONFIG_FILE', type=str, required=True, help='The path to the JSON configuration file.')
parser.add_argument('--OVERPROV_PERC', type=int, required=True, help='The percentage of GPU memory overprovisioning.')
parser.add_argument('--TEMPLATE', type=str, choices=['ALIBABA', 'CUSTSCHED', 'DEFAULT'], default='ALIBABA', help='The YAML template that is going to be used.')

args = parser.parse_args()

OUTPUT_DIR = args.OUTPUT_DIR
# print(OUTPUT_DIR)
CONFIG_FILE = args.CONFIG_FILE
# print(CONFIG_FILE)
OVERPROV_PERC = args.OVERPROV_PERC
# print(OVERPROV_PERC)
TEMPLATE = args.TEMPLATE
# print(TEMPLATE)

##### Create Directory #####
mkdir_p(OUTPUT_DIR)

##### Load JSON Configurations #####
config = open(CONFIG_FILE)
data = json.load(config)

TEMPLATE_NUM = len(data)

##### YAML Creation #####
for i in range(TEMPLATE_NUM):
	MODEL_DIR         = data[i]["MODEL_DIR"]
	DATA_DIR          = data[i]["DATA_DIR"]
	backend           = data[i]["backend"]
	model             = data[i]["model"]
	scenario          = data[i]["scenario"]
	container_name    = data[i]["container_name"]
	image_name        = data[i]["image_name"]
	memory_request_GB = data[i]["memory_request_GB"]
	List		  = data[i]["list"]

	LIST_ITEM_NUM = len(List)   

	for j in range(LIST_ITEM_NUM):
		queries  = List[j]["queries"]
		replicas = List[j]["replicas"]

		parts    = MODEL_DIR.split('/')
		name     = parts[3] + '-' + str(queries) + '-'
		command  = 'cp /root/configs/' + str(queries) + '/mlperf.conf /root/inference/v0.5/ && cd /root/inference/v0.5/classification_and_detection \
&& MODEL_DIR=' + MODEL_DIR + ' DATA_DIR=' + DATA_DIR + ' ./run_local.sh ' + backend + ' ' + model + ' gpu --scenario ' + scenario

		for k in range(replicas):
			f = open(OUTPUT_DIR + '/' + name + str(k) + '.yaml' , "w+")

			out = yaml.dump({})

			if   TEMPLATE == 'ALIBABA':
				out = alibabaTemplate(name + str(k), container_name, image_name, command, getOverprovisionedMemory(memory_request_GB, OVERPROV_PERC))
			elif TEMPLATE == 'CUSTSCHED':
				out = custschedTemplate(name + str(k), container_name, image_name, command, getOverprovisionedMemory(memory_request_GB, OVERPROV_PERC))
			else:
				out = defaultTemplate(name + str(k), container_name, image_name, command)

			f.write(out)

			f.close()

