# syntax=docker/dockerfile:1

FROM golang:1.21 as build
WORKDIR /app
# Download Go modules
COPY go.mod ./
RUN go mod download
# Copy the source code. Note the slash at the end, as explained in https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./
RUN  GOOS=linux go build -o /InstallRootCaCerts

FROM gcr.io/distroless/base-debian11 AS release
WORKDIR /
COPY --from=build /InstallRootCaCerts /InstallRootCaCerts
ENTRYPOINT ["/InstallRootCaCerts"]

