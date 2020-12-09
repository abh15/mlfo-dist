FROM golang:1.15-alpine

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/abh15/mlfo-dist
# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# Run the executable
CMD ["mlfo"]

