FROM --platform=$BUILDPLATFORM golang:1.20 AS build-stage
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY src/commons src/commons
COPY src/controllers src/controllers
COPY src/main.go src/main.go
COPY src/utils src/utils
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /calmOrchestrator src/main.go

FROM alpine:3.14 AS build-release-stage

WORKDIR /

COPY --from=build-stage /calmOrchestrator /calmOrchestrator

ENTRYPOINT ["/bin/sh", "-c","./calmOrchestrator"]
