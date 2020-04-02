import os
import yaml
from datetime import timedelta

from kubernetes import client, config

class Job:
	# Job class constructor
	def __init__ (self, YAML_path):
		self.__YAML_path = YAML_path
		self.__is_applied = False		

		# Parses Job name and aliyun.com/gpu-mem
		with open(YAML_path, 'r') as stream:
			try:
				parsed_YAML = yaml.safe_load(stream)
				self.__name = parsed_YAML['metadata']['name']
			except yaml.YAMLError as e:
				print(e)
		
		# Kubernete client setup
		config.load_kube_config()
		self.__api_instance = client.BatchV1Api()
		

	# Creates the associated Jod
	def apply_job (self):
		command = 'kubectl create -f ' + self.__YAML_path
		os.system(command)
		self.__is_applied = True

	# Deletes the associated Job
	def delete_job (self):
		command = 'kubectl delete -f ' + self.__YAML_path
		os.system(command)

	# Returns whether a Job is completed
	def is_completed (self):
		api_response = self.__api_instance.read_namespaced_job(self.__name, 'default')

		start_time = api_response.status.start_time
		completion_time = api_response.status.completion_time

		if completion_time == None:
			return False
		else:
			self.__execution_time = completion_time - start_time
			
			return True

	# Returns the name
	def get_name (self):
		return self.__name	

	# Returns the execution_time
	def get_execution_time (self):
		return self.__execution_time

	# Returns True if the Job is applied
	def is_applied (self):
		return self.__is_applied

	# Return if the job had QoS violation
	def has_QoS_violation (self):
		delta = timedelta(days=0, seconds=0, microseconds=0, milliseconds=0, minutes=0, hours=0, weeks=0)
		return (delta < self.__execution_time)
