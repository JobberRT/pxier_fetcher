FROM golang:1.18 AS build
WORKDIR /pxier_fetcher
COPY . .
RUN go mod tidy &&  \
    go mod vendor && \
    go build -o pxier_fetcher && \
    cp config.example.yaml config.yaml

FROM ubuntu:22.04 AS run
COPY --from=build /pxier_fetcher/pxier_fetcher .
COPY --from=build /pxier_fetcher/config.yaml .
CMD ["./pxier_fetcher"]