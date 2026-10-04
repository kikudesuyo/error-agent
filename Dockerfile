FROM golang:1.27-alpine AS build

WORKDIR /src
COPY api/go.mod api/go.sum ./api/
RUN cd api && go mod download
COPY api ./api
RUN cd api && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/error-agent .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/error-agent /error-agent
EXPOSE 8080
ENTRYPOINT ["/error-agent"]
