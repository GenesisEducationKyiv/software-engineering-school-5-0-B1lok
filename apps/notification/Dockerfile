ARG GO_VERSION=1.24.3
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

COPY apps/notification/go.mod apps/notification/go.sum ./apps/notification/
COPY pkg/ ./pkg/

RUN --mount=type=cache,target=/go/pkg/mod/ \
    cd apps/notification && go mod download -x

COPY apps/notification/ ./apps/notification/

ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod/ \
    cd apps/notification && CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./cmd/http-server

FROM alpine:3.19 AS final

COPY apps/notification/templates /templates
COPY apps/notification/migrations /migrations

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
