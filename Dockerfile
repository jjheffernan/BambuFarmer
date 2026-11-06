# Multi-service Dockerfile. Build with:
#   docker build --build-arg SERVICE=gateway -t bambufarmer/gateway:dev .
ARG SERVICE=gateway

FROM golang:1.23-bookworm AS build
ARG SERVICE
WORKDIR /src
# Use per-module replace directives; ignore repo go.work inside the image.
ENV GOWORK=off
COPY packages ./packages
COPY services ./services
WORKDIR /src/services/${SERVICE}
RUN go mod tidy && CGO_ENABLED=0 go build -o /out/service .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/service /service
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/service"]
