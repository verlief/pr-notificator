FROM mirror.gcr.io/golang:alpine AS builder

RUN apk add --no-cache tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -buildvcs=false -o main cmd/main.go

EXPOSE 80

FROM mirror.gcr.io/alpine

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/main /app/

CMD ["./main"]
