FROM golang:1.24rc1-alpine
ENV GOTOOLCHAIN=auto
WORKDIR /app
COPY . .
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/main.go
RUN go build -o health-checker ./cmd/main.go
EXPOSE 8080
CMD ["./health-checker"]