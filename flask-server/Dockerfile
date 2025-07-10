FROM python:3.12-slim

WORKDIR /app  

COPY requirements.txt .
COPY flask-server/ .

RUN pip install -r requirements.txt

RUN apt-get update && apt-get install -y curl && \
    pip install -r requirements.txt && \
    curl -sSL https://raw.githubusercontent.com/sqlmapproject/sqlmap/master/sqlmap.py -o /usr/local/bin/sqlmap && \
    chmod +x /usr/local/bin/sqlmap

EXPOSE 5001

CMD ["python", "run.py"]