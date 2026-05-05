FROM golang:1.25.7-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o monopoly-backend .

FROM alpine:3.22
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/monopoly-backend ./monopoly-backend
ENV PORT=9876
EXPOSE 9876
CMD ["./monopoly-backend"]
