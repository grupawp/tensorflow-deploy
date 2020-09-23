FROM jupyter/base-notebook

RUN pip install -U pip setuptools

COPY setup.py /tmp
RUN cd /tmp && python setup.py install

RUN pip install tensorflow-cpu

RUN pip install -U \
	tensorflow-cpu \
	tensorflow-hub \
	tensorflow-serving-api

RUN pip install -U \
	coverage \
	ipywidgets \
	matplotlib \
	pandas \
	requests-mock \
	tensorflow-text

RUN pip install tensorflow-deploy-utils==1

RUN mv /home/jovyan/work /home/jovyan/.work
