FROM golang:1.20 as builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY Makefile ./

COPY . .
RUN make

FROM builder as test
CMD ["make", "test"]

FROM scratch

COPY --from=builder /build/bin/skiperator ./

USER 65532:65532
ENTRYPOINT ["/skiperator"]
