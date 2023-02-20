FROM python:3.10


WORKDIR /app

COPY ./nsfw-classification .

RUN pip install -r requirements.txt

CMD ["uvicorn", "server:app", "--host", "0.0.0.0", "--port", "8000"]