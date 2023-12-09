# Use the official Golang image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code and go.mod/go.sum files into the container
COPY go.mod .
COPY go.sum .
COPY . .

# Install necessary packages
RUN go mod download

# Build the Go application
RUN go build -o main .

# Expose the port the application runs on
EXPOSE 8000

# Command to run the executable
CMD ["./main"]
