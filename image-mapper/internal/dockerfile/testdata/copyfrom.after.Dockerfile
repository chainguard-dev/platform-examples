FROM cgr.dev/chainguard/python:3.13-dev AS python

WORKDIR /app

COPY requirements.txt

RUN pip install --no-cache-dir --target /app -r requirements.txt

FROM cgr.dev/chainguard/python:3.13-dev

COPY --from=python /app /app
COPY --from=cgr.dev/chainguard/python:3.13-dev /etc/example /example

WORKDIR /app

COPY run.py run.py

ENTRYPOINT ["python", "/app/run.py"]
