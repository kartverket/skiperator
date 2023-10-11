FROM golang:1.21-alpine as builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY Makefile ./
COPY . .

RUN apk update && apk add --no-cache bash && apk add --no-cache make
RUN make

FROM builder as test
CMD ["make", "test"]

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/bin/skiperator ./

USER 65532:65532
ENTRYPOINT ["/skiperator"]
