# Benchmark Function Info

## Input
- Input must be in JSON Format of {string:int}
- Parameter names are 'executiontime', 'objectsize', 'memoryallocate'
- Unit for 'executiontime' parameter is miliseconds; 'objectsize' is kB; 'memoryallocate' is kB
- One or more of the three parameters can be utilized. However, there are assumptions
  - if 'executiontime' parameter is used concurrently with other parameters, the 'executiontime' input must be at least enough to complete the other jobs, as it is the total runtime
  - if 'objectsize' parameter is used, the program assumes there's already a object of inputted size in /mybucket, else invoke the create_objectcreate_object_of_size_and_store_in_bucket.sh to creat one
