FROM golang:1.21-alpine
WORKDIR /app
RUN go install github.com/cosmtrek/air@v1.49.0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
EXPOSE 8080
CMD ["air"]