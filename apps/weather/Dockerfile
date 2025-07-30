ARG GO_VERSION=1.24.3
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

COPY apps/weather/go.mod apps/weather/go.sum ./apps/weather/
COPY pkg/ ./pkg/

RUN --mount=type=cache,target=/go/pkg/mod/ \
    cd apps/weather && go mod download -x

COPY apps/weather/ ./apps/weather/

ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod/ \
    cd apps/weather && CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./cmd/grpc-server

FROM alpine:latest AS final

RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser

RUN mkdir /logs && chown -R appuser:appuser /logs

USER appuser

COPY --from=build /bin/server /bin/

EXPOSE 50051

ENTRYPOINT [ "/bin/server" ]
