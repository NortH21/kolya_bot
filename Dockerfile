FROM golang:latest as builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o bot -ldflags '-linkmode external -w -extldflags "-static"' .


FROM alpine:latest  

RUN apk --no-cache add ca-certificates curl
WORKDIR /app
COPY --from=builder /build/bot .
COPY ./files ./files/

USER 9000:9000

CMD ["/app/bot"]

