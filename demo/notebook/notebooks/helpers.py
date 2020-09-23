from google.protobuf.wrappers_pb2 import Int64Value
import grpc
import matplotlib.pyplot as plt
import tensorflow as tf
from tensorflow_serving.apis import predict_pb2
from tensorflow_serving.apis import prediction_service_pb2_grpc
from tqdm import tqdm_notebook


def grpc_client(dataset, host, port, model_name, model_label, model_version=0):
    channel = grpc.insecure_channel(f'{host}:{port}')
    stub = prediction_service_pb2_grpc.PredictionServiceStub(channel)

    results = []
    for d in tqdm_notebook(dataset):
        request = predict_pb2.PredictRequest()
        request.model_spec.name = model_name
        request.model_spec.signature_name = "serving_default"
        if not model_version:
            request.model_spec.version_label = model_label
        else:
            tmp_version = Int64Value()
            tmp_version.value = model_version
            request.model_spec.version.CopyFrom(tmp_version)
        request.inputs["dense_input"].CopyFrom(tf.make_tensor_proto(d, shape=(1, 28, 28), dtype="float32"))
        results.append(stub.Predict(request, timeout=10))
    
    return results


def grpc_text_client(dataset, host, port, model_name, model_label, model_version=0):
    channel = grpc.insecure_channel(f'{host}:{port}')
    stub = prediction_service_pb2_grpc.PredictionServiceStub(channel)

    results = []
    for d in tqdm_notebook(dataset):
        request = predict_pb2.PredictRequest()
        request.model_spec.name = model_name
        request.model_spec.signature_name = "serving_default"
        if not model_version:
            request.model_spec.version_label = model_label
        else:
            tmp_version = Int64Value()
            tmp_version.value = model_version
            request.model_spec.version.CopyFrom(tmp_version)
        request.inputs["x"].CopyFrom(tf.make_tensor_proto(d, shape=None, dtype="string"))
        results.append(stub.Predict(request, timeout=10))
    
    return results


def return_mistakes(data, true_labels, predictions, mistakes_number=0):

    dense_labels = tf.argmax(true_labels, -1).numpy()
    dense_predictions = tf.argmax(predictions, -1).numpy()

    mistakes_indices = []
    mistakes_cnt = 0
    for i, (true, pred) in enumerate(zip(dense_labels, dense_predictions)):
        if true != pred:
            mistakes_indices.append((data[i], true, pred))
            mistakes_cnt += 1
            if mistakes_number and mistakes_cnt >= mistakes_number:
                break
    return mistakes_indices


def show(array, title):
    plt.figure()
    plt.imshow(array)
    plt.axis('off')
    plt.title('\n\n{}'.format(title), fontdict={'size': 16})
    plt.show()


def show_label(array):
    return tf.argmax(array, -1).numpy()
