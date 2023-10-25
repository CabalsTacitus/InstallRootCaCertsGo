# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.21-alpine as build
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY go.mod ./
COPY *.go ./
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /InstallRootCaCerts

FROM scratch AS release
WORKDIR /
COPY --from=build /InstallRootCaCerts /InstallRootCaCerts
ENTRYPOINT ["/InstallRootCaCerts"]

