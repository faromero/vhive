import os
import sys
import glob
import grpc
import json
import argparse
import pandas as pd

import helloworld_pb2
import helloworld_pb2_grpc

# Endpoint file (used when traces are specified)
endpoint_file = 'benchmark_endpoints.txt'

# Limit traces to particular day(s) (between 1 to 14)
# If set to empty list, will process all
filter_days = [1]

def getArgs():
  parser = argparse.ArgumentParser()
  parser.add_argument('--endpoint', '-e', type=str, required=True,
                      dest='endpoint', help='Endpoint to query')
  parser.add_argument('--tracedir', '-d', type=str, required=False,
                      dest='tracedir', default=None,
                      help='Path to traces directory. If not specified, will use manually-inputted parameters.')
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
  for l in lines:
    endpoints_list.append(l.strip())
  fd.close()

  return endpoints_list

def parseTraces(tracedir):
  # Helper function to check whether a trace should be parsed
  def shouldParse(trace_name):
    if not filter_days:
      return True
    for d in filter_days:
      curr_day = 'd%02d' % d
      if curr_day in trace_name:
        return True
    return False

  all_traces = glob.glob(os.path.join(tracedir, '*.csv'))

  # Key: Day (in the form dXX); Value: DataFrame
  invocations_map = {}
  memory_map = {}
  duration_map = {}
  for at in all_traces:
    # Check whether we should parse the next trace
    if not shouldParse(at):
      continue

    print('Now parsing:', os.path.basename(at))

    day = at.split('.')[-2]

    # Read into DF
    df = pd.read_csv(at)

    if 'app_memory' in at:
      # Keep columns: HashApp, AverageAllocatedMb_pct99
      df = df[['HashApp','AverageAllocatedMb_pct99']]
      memory_map[day] = df
    elif 'function_durations' in at:
      # Keep columns: HashApp, HashFunction, percentile_Average_99
      df = df[['HashApp', 'HashFunction', 'percentile_Average_99']]
      duration_map[day] = df
    elif 'invocations_per_function' in at:
      # (Easier to drop.) Drop columns: HashOwner, Trigger
      df = df.drop(['HashOwner', 'Trigger'], axis=1)
      invocations_map[day] = df

  return invocations_map, memory_map, duration_map

# Set any of executiontime, objectsize, or memoryallocate to 0 to skip
def queryFunction(endpoint, executiontime, objectsize, memoryallocate):
  with grpc.insecure_channel(endpoint) as channel:
    stub = helloworld_pb2_grpc.GreeterStub(channel)
    inputjson = {} # Json object to store input

    if executiontime > 0:
      inputjson['executiontime'] = executiontime

    if objectsize > 0:
      inputjson['objectsize'] = objectsize 

    if memoryallocate > 0:
      inputjson['memoryallocate'] = memoryallocate 
     
    input_str = json.dumps(inputjson)
    print('Querying for: %s' % input_str)
    response = stub.SayHello(helloworld_pb2.HelloRequest(name=input_str))
    print(response)

def runExperiment(endpoints, invocations_map, memory_map, duration_map):
  # Helper function to map functions to endpoints, and map functions to app hash
  # The latter is to speed up searches for memory
  def mapEndpoints(endpoints, duration_map):
    endpoint_map = {}
    func_app_map = {}
    for _, v in duration_map.items():
      for _, row in v.iterrows():
        hash_app = row['HashApp']
        hash_function = row['HashFunction']
        if len(endpoints) > 0:
          next_endpoint = endpoints[0]
          endpoints.pop(0)
          endpoint_map[hash_function] = next_endpoint
          func_app_map[hash_function] = hash_app
        else:
          break
      if len(endpoints) == 0:
        break

    return endpoint_map, func_app_map

  endpoint_map, func_app_map = mapEndpoints(endpoints, duration_map)
  print(endpoint_map, func_app_map)
  return

def main(args):
  tracedir = args.tracedir

  if tracedir is not None:
    endpoints = readEndpoints()
    invocations_map, memory_map, duration_map = parseTraces(tracedir)
    runExperiment(endpoints, invocations_map, memory_map, duration_map)
  else:
    endpoint = args.endpoint
    executiontime = args.executiontime
    objectsize = args.objectsize
    memoryallocate = args.memoryallocate
    queryFunction(endpoint, executiontime, objectsize, memoryallocate)

if __name__ == '__main__':
  main(getArgs())

