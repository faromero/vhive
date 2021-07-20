from concurrent import futures
import logging
import os
import grpc

from PIL import Image, ImageOps

import helloworld_pb2
import helloworld_pb2_grpc

from minio import Minio

minioEnvKey = "http://10.138.0.27:9000"
image_name = 'img2.jpeg'
image2_name = 'img3.jpeg'
image_path = 'mybucket/' + image_name
image_path2 = 'mybucket/' +image2_name

responses = ["record_response", "replay_response"]

#minioAddress = os.getenv(minioEnvKey)
minioAddress = "10.138.0.27:9000"

class Greeter(helloworld_pb2_grpc.GreeterServicer):

    def SayHello(self, request, context):
        if minioAddress == None:
            print("NONE!!!!!")
            return None
        else:
            print(minioAddress)

        minioClient = Minio(minioAddress,
                access_key='minioadmin',
                secret_key='minioadmin',
                secure=False)
        if minioClient.bucket_exists('mybucket'):
            print("Bucket found!")
            objects = minioClient.list_objects('mybucket', recursive=True)
            for obj in objects:
                print(obj)
        else:
            print("Bucket NOT found!")

        if request.name == "record":
            msg = 'Hello, %s!' % responses[0]
            minioClient.fget_object('mybucket', image_name, image_path)
            image = Image.open(image_path)
            img = image.transpose(Image.ROTATE_90)
        elif request.name == "replay":
            msg = 'Hello, %s!' % responses[1]
            minioClient.fget_object('mybucket', image2_name, image_path2)
            image2 = Image.open(image_path2)
            img = image2.transpose(Image.ROTATE_90)
        else:
            msg = 'Hello, %s!' % request.name
            minioClient.fget_object('mybucket', image_name, image_path)
            image = Image.open(image_path)
            img = image.transpose(Image.ROTATE_90)

        return helloworld_pb2.HelloReply(message=msg)


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    helloworld_pb2_grpc.add_GreeterServicer_to_server(Greeter(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    server.wait_for_termination()


if __name__ == '__main__':
    logging.basicConfig()
    serve()
