
FROM golang:1.19 AS builder

WORKDIR /app

COPY "go.mod" "go.sum" ./

RUN go mod download

COPY . .

RUN \
    make test && \
	make


FROM alpine:3

RUN \
	apk update && \
	apk add ca-certificates libc6-compat && \
	rm -rf /var/cache/apk/*

COPY --from=builder /app/build/nsdns /usr/bin/

CMD ["/usr/bin/nsdns"]
