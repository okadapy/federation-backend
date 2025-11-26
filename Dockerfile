FROM golang:latest
LABEL authors="OKADA"


WORKDIR /app
COPY . ./ 
RUN go run .

