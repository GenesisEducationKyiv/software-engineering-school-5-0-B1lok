ARG GO_VERSION=1.24.3
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

COPY apps/gateway/go.mod apps/gateway/go.sum ./apps/gateway/

RUN --mount=type=cache,target=/go/pkg/mod/ \
    cd apps/gateway && go mod download -x

COPY apps/gateway/ ./apps/gateway/

ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod/ \
    cd apps/gateway && CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./cmd/http-server

FROM alpine:3.21 AS final

COPY apps/gateway/templates /templates

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

USER appuser

COPY --from=build /bin/server /bin/

EXPOSE 8080

ENTRYPOINT [ "/bin/server" ]
