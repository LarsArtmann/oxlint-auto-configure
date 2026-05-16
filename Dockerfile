FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build     -tags netgo     -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}"     -trimpath     -o oxlint-auto-configure     ./cmd/oxlint-auto-configure

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /build/oxlint-auto-configure /oxlint-auto-configure

USER 65532:65532

ENTRYPOINT ["/oxlint-auto-configure"]
