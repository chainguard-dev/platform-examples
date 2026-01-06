ARG IMAGE_REGISTRY
ARG IMAGE=python

FROM ${IMAGE} as latest
FROM ${IGNORED} as ignored
FROM ${IMAGE}:3.13 AS python

ARG IMAGE_REGISTRY=ignored

COPY requirements.txt

RUN pip install --no-cache-dir --target /app -r requirements.txt

FROM ${IMAGE_REGISTRY:-docker.io}/python

WORKDIR /app

COPY run.py run.py

ENTRYPOINT ["python", "/app/run.py"]
