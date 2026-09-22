FROM gcr.io/distroless/static-debian13:nonroot

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/oxlint-auto-configure /oxlint-auto-configure

USER 65532:65532

ENTRYPOINT ["/oxlint-auto-configure"]
