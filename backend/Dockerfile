# Build
FROM golang:1.27.3-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8888

CMD ["go", "run", "."]