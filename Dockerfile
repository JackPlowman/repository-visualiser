#checkov:skip=CKV_DOCKER_2
#checkov:skip=CKV_DOCKER_3
FROM golang:1.23.5-bookworm

# Install Git
RUN apt-get update && apt-get install -y git

# Configure Git
RUN git config --global user.name "github-actions" && \
    git config --global user.email "github-actions@github.com"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go ignore.json ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /application

CMD ["/application"]
