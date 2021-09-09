import grpc
import json
import logging
import argparse
import helloworld_pb2
import helloworld_pb2_grpc

def SayHello(address):
    parser = argparse.ArgumentParser()
    parser.add_argument('--time', '-t', help='input desired execution time in uni\
t miliseconds', type= int, dest = 'executiontime', required = False)
    parser.add_argument('--size','-s',  help = 'input desired object size in unit\
 of kB', type = int, dest = 'objectsize', required=False)
    parser.add_argument('--mem','-m', help = 'input desired memory allocation siz\
e in unit of kB', type = int, dest='memoryallocate',  required= False)
    args = parser.parse_args()
    print(type(args))

    runDuration = 5; # Run the experiment for 5 seconds                           
    targetRPS = 1.0 # Target requests per second                                  


    with grpc.insecure_channel(address) as channel:

        stub = helloworld_pb2_grpc.GreeterStub(channel)
        userinput = {}
        input_exists = False

        if args.executiontime:
            t_json = {'executiontime': args.executiontime}
            userinput.update(t_json)
            input_exists = True

        if args.objectsize:
            s_json = {'objectsize': args.objectsize}
            userinput.update(s_json)
            input_exist = True
            
        if args.memoryallocate:
            m_json = {'memoryallocate': args.memoryallocate}
            userinput.update(m_json)
            input_exists = True
        if input_exists == False:
            print('Error Message: Please input at least one parameter to be bench\
mark-simulated. Type with -h flag to see parameter input flags')
            return
        input_str = json.dumps(userinput)
        print('sending string input: ' + input_str)
        response = stub.SayHello(helloworld_pb2.HelloRequest(name=input_str))
        print('SayHello Completed')
        print(str(response))
def readEndpoints():

    with open('/home/yalew/vhive/endpoints.json',"r" ) as json_file:
        data = json.load(json_file)[0]
        return data
def mainfunc():
    parser = argparse.ArgumentParser()
    parser.add_argument('executiontime', help='input desired execution time in un\
it miliseconds', type= int)
    parser.add_argument('objectsize', help = 'input desired object size in unit o\
f kB', type = int)
    parser.add_argument('memoryallocate', help = 'input desired memory allocation\
 size in unit of kB', type = int)
    endpointsinfo = readEndpoints()
    address = endpointsinfo['hostname']
    SayHello(address)

if __name__ == '__main__':
    logging.basicConfig()
    mainfunc()

