FROM python:3.13 AS python

WORKDIR /app

COPY requirements.txt

RUN --mount=type=bind,from=python,target=/etc/example \
    --mount=type=cache,target=/etc/pip,from=python:3.13 \
    pip install --no-cache-dir --target /app -r requirements.txt \
    && rm requirements.txt

FROM python

WORKDIR /app

COPY run.py run.py

RUN --mount=type=bind,from=docker.io/python,target=/bin/cat \
     cat run.py


ENTRYPOINT ["python", "/app/run.py"]
