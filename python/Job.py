import os
import yaml

from datetime import timedelta
from kubernetes import client, config
from kubernetes.client.rest import ApiException

class Job:
	# Job class constructor
	def __init__ (self, YAML_path):
		# YAML Path
		self.__YAML_path                 = YAML_path
		# Job Fields
		self.__Job_name                  = None
		self.__Job_is_applied            = False
		self.__Job_creation_timestamp    = None
		self.__Job_start_time            = None
		self.__Job_completion_time       = None
		# Pod Fields
		self.__Pod_name                  = None
		self.__Pod_creation_timestamp    = None
		self.__Pod_start_time            = None
		self.__Pod_container_started_at  = None
		self.__Pod_container_finished_at = None
		self.__Pod_container_logs        = None
		self.__Pod_container_90percile   = None
		self.__Pod_container_99percile   = None

		# Parse YAML file (Job_name)
		self.__parse_YAML(YAML_path)

		# Kubernete client setup
		config.load_kube_config()
		self.__BatchV1API_instance = client.BatchV1Api()
		self.__CoreV1API_instance  = client.CoreV1Api()

	# Parses the YAML file
	def __parse_YAML (self, YAML_path):
		with open(YAML_path, 'r') as stream:
			try:
				parsed_YAML = yaml.safe_load(stream)
				# Get Job_name
				self.__Job_name = parsed_YAML['metadata']['name']
 			except yaml.YAMLError as e:
				print(e)

	# Creates the Jod
	def apply_job (self):
		command = 'kubectl create -f ' + self.__YAML_path
		os.system(command)

		self.__Job_is_applied = True

	# Deletes the Job
	def delete_job (self):
		command = 'kubectl delete -f ' + self.__YAML_path
		os.system(command)
	
	# Execute k8s API calls
	def __exec_k8s_API_c (self):
		try:
			BatchV1API_response = self.__BatchV1API_instance.read_namespaced_job(self.__Job_name, 'default')

			# print(BatchV1API_response)

			self.__Job_creation_timestamp = BatchV1API_response.metadata.creation_timestamp
			self.__Job_start_time         = BatchV1API_response.status.start_time
			self.__Job_completion_time    = BatchV1API_response.status.completion_time
		except ApiException as e:
			print("Exception when calling BatchV1API->read_namespaced_job: %s\n" % e)

		if self.__Pod_name == None:
			try:
				CoreV1API_response = self.__CoreV1API_instance.list_namespaced_pod('default')

				List = CoreV1API_response.items
				for index in range(len(List)):
					p_name = List[index].metadata.name
					if self.__Job_name in p_name:
						self.__Pod_name = p_name
						break
			except ApiException as e:
				print("Exception when calling CoreV1API->list_namespaced_pod: %s\n" % e)
		else:
			try:
				CoreV1API_response = self.__CoreV1API_instance.read_namespaced_pod(self.__Pod_name, 'default')

				# print(CoreV1API_response)

				self.__Pod_creation_timestamp = CoreV1API_response.metadata.creation_timestamp
				self.__Pod_start_time         = CoreV1API_response.status.start_time
				if CoreV1API_response.status.container_statuses != None and CoreV1API_response.status.container_statuses[0].state.terminated != None:
					self.__Pod_container_started_at  = CoreV1API_response.status.container_statuses[0].state.terminated.started_at
					self.__Pod_container_finished_at = CoreV1API_response.status.container_statuses[0].state.terminated.finished_at
			except ApiException as e:
				print("Exception when calling CoreV1API->read_namespaced_pod: %s\n" % e)

			try:
				CoreV1API_response = self.__CoreV1API_instance.read_namespaced_pod_log(self.__Pod_name, 'default', tail_lines=1)
				
				self.__Pod_container_logs = CoreV1API_response

			except ApiException as e:
				pass
				# print("Exception when calling CoreV1API->read_namespaced_pod_log: %s\n" % e)


	# Returns whether a Job is completed
	def is_completed (self):
		self.__exec_k8s_API_c()

		# self.__get_state()

		return (self.__Job_completion_time != None)

	# Returns the Job name
	def get_name (self):
		return self.__Job_name	

	# Returns the Job execution time
	def get_Job_execution_time (self):
		return self.__Job_completion_time - self.__Job_start_time

	# Returns the Container execution time
	def get_Container_execution_time (self):
		return self.__Pod_container_finished_at - self.__Pod_container_started_at

	def get_Job_stats (self):
		parts = self.__Pod_container_logs.split(',')
				
		self.__Pod_container_90percile = parts[6].split(':')[1]
		self.__Pod_container_99percile = parts[8].split(':')[1]

		return  """Job Name                                     = """ + self.__Job_name + '\n' + \
			"""Job_Start_Time - Job_Creation_Timestamp      = """ + str(self.__Job_start_time - self.__Job_creation_timestamp) + '\n' + \
			"""Pod_Creation_Timestamp - Job_Start_Time      = """ + str(self.__Pod_creation_timestamp - self.__Job_start_time) + '\n' + \
			"""Pod_Start_Time - Pod_Creation_Timestamp      = """ + str(self.__Pod_start_time - self.__Pod_creation_timestamp) + '\n' + \
			"""Container_Started_At - Pod_Start_Time        = """ + str(self.__Pod_container_started_at - self.__Pod_start_time) + '\n' + \
			"""Container_Finished_At - Container_Started_At = """ + str(self.__Pod_container_finished_at - self.__Pod_container_started_at) + '\n' + \
			"""Job_Completion_Time - Container_Finished_At  = """ + str(self.__Job_completion_time - self.__Pod_container_finished_at) + '\n' + \
			"""Job Execution Time                           = """ + str(self.__Job_completion_time - self.__Job_start_time) + '\n' + \
			"""Container 90%-ile                            = """ + str(self.__Pod_container_90percile) + ' msec\n' + \
			"""Container 99%-ile                            = """ + str(self.__Pod_container_99percile) + ' msec\n' + '\n'

	# Returns True if the Job is applied
	def is_applied (self):
		return self.__Job_is_applied
	
	# Returns current Job state
	def __get_state (self):
		print("Job Name                  : %s" % self.__Job_name)
		print("Job Creation Timestamp    : %s" % self.__Job_creation_timestamp)
		print("Job Start Time            : %s" % self.__Job_start_time)
		print("Job Completion Time       : %s" % self.__Job_completion_time)

		print("Pod Name                  : %s" % self.__Pod_name)
		print("Pod Creation Timestamp    : %s" % self.__Pod_creation_timestamp)
		print("Pod Start Time            : %s" % self.__Pod_start_time)
		print("Pod Container Started at  : %s" % self.__Pod_container_started_at)
		print("Pod Container Finished at : %s" % self.__Pod_container_finished_at)
		print("Pod Container 90%-ile     : %s ms" % self.__Pod_container_90percile)
		print("Pod Container 99%-ile     : %s ms" % self.__Pod_container_99percile)

