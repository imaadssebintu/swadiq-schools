FROM golang:1.25-alpine
RUN apk add --no-cache tzdata
ENV TZ=Africa/Nairobi
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
EXPOSE 8080
CMD ["./main"]