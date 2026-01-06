FROM python:3.13

COPY requirements.txt

RUN pip install --no-cache-dir -r requirements.txt

WORKDIR /app

COPY run.py run.py

ENTRYPOINT ["python", "/app/run.py"]
