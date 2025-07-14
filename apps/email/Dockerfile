ARG GO_VERSION=1.24.3
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

COPY apps/email/go.mod apps/email/go.sum ./apps/email/

RUN --mount=type=cache,target=/go/pkg/mod/ \
    cd apps/email && go mod download -x

COPY apps/email/ ./apps/email/

ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod/ \
    cd apps/email && CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./cmd/http-server

FROM alpine:latest AS final

COPY apps/email/templates /templates

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

EXPOSE 8383

ENTRYPOINT [ "/bin/server" ]
