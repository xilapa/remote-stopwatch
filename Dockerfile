# https://chemidy.medium.com/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
FROM golang:alpine as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create alpine user
ENV USER=appuser
ENV UID=10001

# See https://stackoverflow.com/a/55757473/12429735
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /build
COPY . .

# Build the binary
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /bin/remote-stopwatch

# Fetch dependencies.
RUN go get -d -v

## BUILD THE FINAL IMAGE ##
FROM scratch
# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the executable
COPY --from=builder /bin/remote-stopwatch /remote-stopwatch
# Copy the html pages
COPY --from=builder build/pages /pages
# Copy the config yml
COPY --from=builder build/config.yml /config.yml

# Use an unprivileged user.
USER appuser:appuser

# Expose port
ARG PORT
EXPOSE $PORT/tcp
ENV PORT=$PORT

# Run the hello binary.
ENTRYPOINT ["/remote-stopwatch"]
