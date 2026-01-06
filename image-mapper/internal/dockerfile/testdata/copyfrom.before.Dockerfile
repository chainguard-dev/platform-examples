FROM python:3.13 AS python

WORKDIR /app

COPY requirements.txt

RUN pip install --no-cache-dir --target /app -r requirements.txt

FROM python:3.13

COPY --from=python /app /app
COPY --from=python:3.13 /etc/example /example

WORKDIR /app

COPY run.py run.py

ENTRYPOINT ["python", "/app/run.py"]
