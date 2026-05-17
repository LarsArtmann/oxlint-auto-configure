FROM gcr.io/distroless/static-debian13:nonroot

COPY oxlint-auto-configure /oxlint-auto-configure

USER 65532:65532

ENTRYPOINT ["/oxlint-auto-configure"]
