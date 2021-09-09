import grpc
import time
import json
import logging
import os
import minio

from concurrent import futures
from minio.commonconfig import CopySource
from minio import Minio

import helloworld_pb2
import helloworld_pb2_grpc

class Greeter(helloworld_pb2_grpc.GreeterServicer):
  # Get current time in ms
  def _getMilliTime(self):
    return round(time.time() * 1e3)

  # Function to allocate memory
  # Input is memory amount in KB
  # Returns allocated_list (None if failed)
  def _doMemoryAllocation(self, mem_amount_kb):
    # Obtain memory info of the system by looking into the /proc/meminfo folder
    mem_limit = 0
    with open('/proc/meminfo', 'r') as file:
      data = file.read().replace('\n', '')
      data = data.split()
      mem_limit = int(data[data.index('kBMemFree:')+1])
      mem_limit *= 0.125 # Convert to KB

    dummylist = None
    if mem_limit < mem_amount_kb:
      print('Not enough free memory!')
      return dummylist

    # Allocate memory amout specified
    print('Allocating %dKB' % mem_amount_kb)
    mem_amount_b = round(mem_amount_kb*1024)
    dummylist = [0] * mem_amount_b

    return dummylist

  # Function to request from object storage
  # Input is object size in KB
  # Returns object size that was fetched (None if failed)
  def _doStorageRequest(self, object_size_kb):
    print('Fetching an object of size %dKB' % object_size_kb)

    # Connect to minio storage server
    client = Minio("10.138.0.34:9000", access_key="minioadmin", secret_key="minioadmin", secure=False)

    buckets = client.list_buckets()
    objectname = None
    for bucket in buckets:
      if bucket.name == 'mybucket': # Locate desired bucket
        objects = client.list_objects(bucket.name)
        for objs in objects:
          data = client.stat_object(bucket.name, objs.object_name)
          data_size = data.size * 1e3
          if data_size == object_size_kb: #/1000*1024:                 
            objectname = objs.object_name
            print('Desired object found: %s' % objectname)
            break

    if objectname is None:
      print('Object not found')
      return objectname
 
    obj = client.get_object('mybucket', objectname)
    # Download file to /tmp
    with open(os.path.join("/tmp/", objectname), "wb") as tmpfile:
      for d in obj.stream(32*1024):
        tmpfile.write(d)
      print('Saved to /tmp directory')

    obj.close()
    obj.release_conn()

    return objectname

  # Function to "execute" for a set amount of time in ms
  # Input is the amount of time to execute for
  # No return value
  def _doExecutionTime(self, input_time_ms):
    print ('Running for %d ms' % input_time_ms)

    curr_time_ms = self._getMilliTime()
    end_time = curr_time_ms + input_time_ms
    while curr_time_ms < end_time:
      dummyoperation = 1 + 1
      curr_time_ms = self._getMilliTime()

    return

  def SayHello(self, request, context):
    msg = '' # Message to return to client
    userinput = json.loads(request.name)

    memorysize = userinput.get('memoryallocate', None)
    objectsize = userinput.get('objectsize', None)
    executiontime = userinput.get('executiontime', 200)

    memory_allocation_time = 0
    object_fetch_time = 0

    # Check if user asked to do a memory allocation
    if memorysize is not None:
      memory_allocation_timer_start = self._getMilliTime()
      allocated_list = self._doMemoryAllocation(memorysize)

      if allocated_list is None:
        msg = 'Not enough memory on the heap. Try a smaller size.'
        return helloworld_pb2.HelloReply(message=msg)

      memory_allocation_time = self._getMilliTime() - memory_allocation_timer_start
      msg = 'Allocated %dKB in %dms\n' % (memorysize, memory_allocation_time)

    # Check whether user asked to do a remote storage fetch
    if objectsize is not None:
      object_fetch_timer_start = self._getMilliTime()
      objectname = self._doStorageRequest(objectsize)

      if objectname is None:
        msg = msg + 'Object of desired size does not exist in the bucket.\n'
        return helloworld_pb2.HelloReply(message=msg)

      object_fetch_time = self._getMilliTime() - object_fetch_timer_start
      msg = msg + '%dKB fetched to /tmp in %dms\n' % (objectsize, object_fetch_time)

    # Run execution time
    timeleft = executiontime - memory_allocation_time - object_fetch_time

    # Verify the time input is sufficient
    if timeleft < 0:
      msg = msg + 'More time needed for the other benchmark operations.\n'
      return helloworld_pb2.HelloReply(message=msg)

    self._doExecutionTime(timeleft)
    msg = msg + 'Execution time completed for %dms' % executiontime

    print(msg)
    return helloworld_pb2.HelloReply(message=msg)

def serve():
  server = grpc.server(futures.ThreadPoolExecutor(max_workers=5))
  helloworld_pb2_grpc.add_GreeterServicer_to_server(Greeter(), server)
  server.add_insecure_port('[::]:50051')
  server.start()
  server.wait_for_termination()

if __name__ == '__main__':
  logging.basicConfig()
  serve()

