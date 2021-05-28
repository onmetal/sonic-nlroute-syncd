FROM golang:alpine

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o sonic-nlroute-syncd .

WORKDIR /dist

RUN cp /build/sonic-nlroute-syncd .

CMD ["/dist/sonic-nlroute-syncd"]