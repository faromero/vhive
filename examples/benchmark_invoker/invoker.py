import os
import sys
import grpc
import json
import time
import argparse
from threading import *
from timeit import default_timer as now
from multiprocessing import Process, Manager
from trace_manager import TraceQuery, TraceManager

from kubernetes import config, client
from kubernetes.client import Configuration
from kubernetes.client.api import core_v1_api

import helloworld_pb2
import helloworld_pb2_grpc

# Endpoint file (used when traces are specified)
endpoint_file = 'benchmark_endpoints.txt'

# Limit traces to particular day(s) (between 1 to 14)
# If set to empty list, will process all
filter_days = [1]

def getArgs():
  parser = argparse.ArgumentParser()
  parser.add_argument('--endpoint', '-e', type=str, required=False,
                      dest='endpoint', default='http://localhost:80',
                      help='Endpoint to query')
  parser.add_argument('--tracedir', '-d', type=str, required=False,
                      dest='tracedir', default=None,
                      help='Path to traces directory. If not specified, will use manually-inputted parameters.')
  parser.add_argument('--minrange', '-q', type=int, required=False,
                      dest='minrange', default=1,
                      help='Minimum range to map traces to (Default: 1)')
  parser.add_argument('--maxrange', '-r', type=int, required=False,
                      dest='maxrange', default=10,
                      help='Maximum range to map traces to (Default: 10)')
  parser.add_argument('--numsec', '-s', type=int, required=False,
                      dest='numsec', default=5,
                      help='Number of seconds to simulate. -1 for all (Default: 5)')
  parser.add_argument('--sorttraces', '-u', type=str, required=False,
                      dest='sorttraces', default="default",
                      choices=["mintime", "maxtime", "minmem", "maxmem", "default"],
                      help='Number of seconds to simulate. -1 for all (Default: 5)')
  parser.add_argument('--executiontime', '-t', type=int, required=False,
                      dest='executiontime', default=0,
                      help='Execution time in ms (Default: 0 for skip)')
  parser.add_argument('--objectsize', '-o', type=int, required=False,
                      dest='objectsize', default=0,
                      help='Object size in KB (Default: 0 for skip)')
  parser.add_argument('--memoryallocate', '-m', type=int, required=False,
                      dest='memoryallocate', default=0,
                      help='Memory to allocate in KB (Default: 0 for skip)')
  return parser.parse_args()

# Function for reading in endpoint from file        
def readEndpoints():
  endpoints_list = []
  fd = open(endpoint_file, 'r')
  for l in fd:
    endpoints_list.append(l.strip())
  fd.close()

  return endpoints_list

# Function to monitor how many pods per-function are spun up
# Note that we treat the return_dict as a set
def podMonitorDaemon(return_dict, stop):
  # Get kubernetes client
  config.load_kube_config()
  c = Configuration().get_default_copy()
  Configuration.set_default(c)
  core_v1 = core_v1_api.CoreV1Api()

  while True:
    curr_pods = core_v1.list_namespaced_pod("default")
    for pod in curr_pods.items:
      curr_name = pod.metadata.name

      # Add to dictionary (treatign as a set)
      return_dict[curr_name] = 1

    if stop():
      break
    time.sleep(1)

# Set any of executiontime, objectsize, or memoryallocate to 0 to skip
def queryFunction(functionname, endpoint, executiontime, objectsize, memoryallocate, returndict):
  with grpc.insecure_channel(endpoint) as channel:
    stub = helloworld_pb2_grpc.GreeterStub(channel)
    inputjson = {} # Json object to store input

    if executiontime > 0:
      inputjson['executiontime'] = int(executiontime)

    if objectsize > 0:
      inputjson['objectsize'] = int(objectsize)

    if memoryallocate > 0:
      inputjson['memoryallocate'] = int(memoryallocate)
     
    input_str = json.dumps(inputjson)
    # print('Querying for: %s' % input_str)

    start = now()
    response = stub.SayHello(helloworld_pb2.HelloRequest(name=input_str))
    end = now()

    e2e_time = (end - start) * 1e3

    # With a manager dict, we cannot simply append to the list from within the dictionary.
    next_list = returndict[functionname]
    next_list.append(e2e_time)
    returndict[functionname] = next_list 

def runExperiment(queries_to_run, num_sec):
  # If num_sec is negative, determine the number of queries to run
  if num_sec < 0:
    num_sec = len(queries_to_run[0].invocations)

  # Track total number of times a function is invoked
  func_invoc_dict = {}

  # Shared dictionaries
  manager = Manager()
  daemon_measurement_dict = manager.dict()
  func_measurement_dict = manager.dict()

  # Launch daemon thread with a stop flag lambda
  stop_flag = False
  daemon = Thread(target=podMonitorDaemon, args=(daemon_measurement_dict, lambda: stop_flag))
  daemon.setDaemon(True)
  daemon.start()

  for i in range(num_sec):
    for q in queries_to_run:
      num_invoc = int(q.invocations[i])

      # Set up both dictionaries at the same time
      if q.function_name not in func_measurement_dict:
        func_measurement_dict[q.function_name] = []
        func_invoc_dict[q.function_name] = 0
      func_invoc_dict[q.function_name] += num_invoc

      processes = [Process(target=queryFunction,
                           args=(q.function_name, q.endpoint, q.execution_time,
                                 q.object_size, q.memory, func_measurement_dict))
                           for x in range(num_invoc)]
      for p in processes:
        p.start()
      for p in processes:
        p.join()

  # Tell the daemon to exit
  stop_flag = True
  daemon.join()

  return func_measurement_dict, daemon_measurement_dict, func_invoc_dict

def main(args):
  tracedir = args.tracedir
  min_range = args.minrange
  max_range = args.maxrange
  num_sec = args.numsec
  sort_traces = args.sorttraces

  if tracedir is not None:
    endpoints = readEndpoints()
    trace_manager = TraceManager(endpoints, tracedir, filter_days, sort_traces)
    queries_to_run = trace_manager.generateQueries(min_range, max_range)
    function_dict, daemon_dict, invoc_dict = runExperiment(queries_to_run, num_sec)
    trace_manager.analyzeResults(function_dict, daemon_dict, invoc_dict)
  else:
    endpoint = args.endpoint
    executiontime = args.executiontime
    objectsize = args.objectsize
    memoryallocate = args.memoryallocate
    queryFunction(endpoint, executiontime, objectsize, memoryallocate)

if __name__ == '__main__':
  main(getArgs())

