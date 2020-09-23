FROM tensorflow/serving:latest

RUN mv /models /models.stock
RUN ln -s /tfdeploy/models /models
ENTRYPOINT ["tensorflow_model_server", "--allow_version_labels_for_unavailable_models=true", "--port=8500", "--rest_api_port=8501", "--model_config_file=/models/empty.config"]
