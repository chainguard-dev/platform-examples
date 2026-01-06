FROM cgr.dev/chainguard/python:3.13-dev AS python

WORKDIR /app

COPY requirements.txt

RUN --mount=type=bind,from=python,target=/etc/example     --mount=type=cache,target=/etc/pip,from=cgr.dev/chainguard/python:3.13-dev     pip install --no-cache-dir --target /app -r requirements.txt     && rm requirements.txt

FROM cgr.dev/chainguard/python:latest-dev

WORKDIR /app

COPY run.py run.py

RUN --mount=type=bind,from=cgr.dev/chainguard/python:latest-dev,target=/bin/cat      cat run.py


ENTRYPOINT ["python", "/app/run.py"]
