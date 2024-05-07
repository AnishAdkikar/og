FROM golang

WORKDIR /db

COPY go.mod go.sum ./
RUN go mod download

COPY * ./
CMD ["go","run","main.go"]