ARG GO_VERSION=1.24.3
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

COPY apps/subscription/go.mod apps/subscription/go.sum ./apps/subscription/
COPY proto-contracts/ ./proto-contracts/

RUN --mount=type=cache,target=/go/pkg/mod/ \
    cd apps/subscription && go mod download -x

COPY apps/subscription/ ./apps/subscription/

ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod/ \
    cd apps/subscription && CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./cmd/grpc-server

FROM alpine:latest AS final

COPY apps/subscription/templates /templates
COPY apps/subscription/migrations /migrations

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

EXPOSE 50052

ENTRYPOINT [ "/bin/server" ]
