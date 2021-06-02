FROM golang as builder
ADD . /go/sonic-nlroute-syncd/
WORKDIR /go/sonic-nlroute-syncd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/sonic-nlroute-syncd

FROM alpine:latest
RUN apk --no-cache add ca-certificates bash
WORKDIR /app
COPY --from=builder /go/bin/sonic-nlroute-syncd .
EXPOSE 9324

ENTRYPOINT ["/app/sonic-nlroute-syncd"]