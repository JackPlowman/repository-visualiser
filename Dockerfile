#checkov:skip=CKV_DOCKER_2
#checkov:skip=CKV_DOCKER_3
FROM golang:1.23.5-bookworm

RUN apt-get update && apt-get install -y \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go ignore.json ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /application

CMD ["/application"]
