FROM --platform=$BUILDPLATFORM golang:1.19-alpine3.16 as build
WORKDIR /src
ARG TARGETOS TARGETARCH
RUN --mount=target=. --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o /out/domain_exporter -ldflags "-s -w"

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out /usr/bin
EXPOSE 9222
ENTRYPOINT ["/usr/bin/domain_exporter"]
