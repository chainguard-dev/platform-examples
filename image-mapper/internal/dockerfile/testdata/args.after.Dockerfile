ARG IMAGE_REGISTRY
ARG IMAGE=python

FROM cgr.dev/chainguard/python:latest-dev as latest
FROM ${IGNORED} as ignored
FROM cgr.dev/chainguard/python:3.13-dev AS python

ARG IMAGE_REGISTRY=ignored

COPY requirements.txt

RUN pip install --no-cache-dir --target /app -r requirements.txt

FROM cgr.dev/chainguard/python:latest-dev

WORKDIR /app

COPY run.py run.py

ENTRYPOINT ["python", "/app/run.py"]
