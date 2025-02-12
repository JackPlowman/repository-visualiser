#checkov:skip=CKV_DOCKER_2
#checkov:skip=CKV_DOCKER_3
FROM golang:1.23.5-bookworm

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY visualiser/*.go visualiser/ignore.json ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /application

CMD ["/application"]
