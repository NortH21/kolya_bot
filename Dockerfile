FROM golang:latest as builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bot -ldflags '-linkmode external -w -extldflags "-static"' .


FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/bot .
COPY ./files .

CMD ["/app/bot"]
