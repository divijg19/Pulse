# syntax=docker/dockerfile:1

# GoReleaser builds the static binary (with the embedded web UI) for each
# target platform and places it in the Docker build context at
# <goos>/<goarch>/pulse. We only COPY the pre-built binary here.
FROM gcr.io/distroless/static-debian12:nonroot
ARG TARGETPLATFORM
WORKDIR /app
COPY $TARGETPLATFORM/pulse /usr/local/bin/pulse
EXPOSE 8080
ENTRYPOINT ["pulse"]
CMD ["web"]
