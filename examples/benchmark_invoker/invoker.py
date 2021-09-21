import os
import sys
import grpc
import json
import argparse
from trace_manager import TraceQuery, TraceManager

import helloworld_pb2
import helloworld_pb2_grpc

# Endpoint file (used when traces are specified)
endpoint_file = 'benchmark_endpoints.txt'

# Limit traces to particular day(s) (between 1 to 14)
# If set to empty list, will process all
filter_days = [1,2]

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

# Set any of executiontime, objectsize, or memoryallocate to 0 to skip
def queryFunction(endpoint, executiontime, objectsize, memoryallocate):
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
    print('Querying for: %s' % input_str)
    response = stub.SayHello(helloworld_pb2.HelloRequest(name=input_str))

    return response

def runExperiment(queries_to_run, num_sec):
  # If num_sec is negative, determine the number of queries to run
  if num_sec < 0:
    num_sec = len(queries_to_run[0].invocations)

  for i in range(num_sec):
    for q in queries_to_run:
      num_invoc = q.invocations[i]
      # TODO: invoke in parallel
      for ni in num_invoc:
        queryFunction(q.endpoint, q.execution_time, q.object_size, q.memory)

  return

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
    runExperiment(queries_to_run, num_sec)
  else:
    endpoint = args.endpoint
    executiontime = args.executiontime
    objectsize = args.objectsize
    memoryallocate = args.memoryallocate
    queryFunction(endpoint, executiontime, objectsize, memoryallocate)

if __name__ == '__main__':
  main(getArgs())

