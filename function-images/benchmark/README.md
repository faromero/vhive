# Benchmark Function Info
This benchmark functions allows the user to simulate behavior of a variety of functions within the vHive environment, and thereby further explore applications. User calls a grpc client and requests from the grpc server the simulation of an application based on one or more of three parameters, 1) runtime, 2) access and downloading of an object of a specific size, 3)allocating memory on the heap. The grpc server completes the request and returns the response and leaves traces to be studied.


## server.py 
backbone of benchmark function, handles requests from grpc clients. 

## greeter_client.py 
grpc client, takes input of a specific format and sends request to the server locally

### Input
- Input must be in JSON Format of **{string:int}**
- Parameter names are **'executiontime'**, **'objectsize'**, **'memoryallocate'**
- Unit for 'executiontime' parameter is **miliseconds**; 'objectsize' is **kB**; 'memoryallocate' is **kB**
- One or more of the three parameters can be utilized. However, there are **assumptions**
  - if 'executiontime' parameter is used concurrently with other parameters, the 'executiontime' input must be at least enough to complete the other jobs, as it is the total runtime
  - if 'objectsize' parameter is used, the program assumes there's already a object of inputted size in /mybucket, else invoke the create_objectcreate_object_of_size_and_store_in_bucket.sh to creat one

## DOCKERFILE 
allows the function to be uploaded to a Docker Container, which is where Vhive calls the server.py

## requirements.txt
contains a list of packages to be installed for the python project to run

## helloworld_pb2_grpc.py/helloworld_pb2.py 
allows the utilization of *[grpc](https://grpc.io/)
