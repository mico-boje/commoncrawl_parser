FROM golang:1.20.1-bullseye

ENV AWS_ACCESS_KEY_ID=EXAMPLE
ENV AWS_SECRET_ACCESS_KEY=EXAMPLEKEY

RUN apt update -y

RUN apt-get install curl -y

RUN apt-get install unzip -y

RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"

RUN unzip "awscliv2.zip"

RUN ./aws/install

RUN aws configure set region eu-west-1

WORKDIR /app

COPY ./scraper .

ENV PATH="${PATH}:/usr/local/go/bin"

CMD ["go", "run", "main.go", "1000"]