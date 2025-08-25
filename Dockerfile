# Multi-stage Dockerfile for QuikGit
FROM alpine:latest

# Install required dependencies
RUN apk add --no-cache ca-certificates git

# Create non-root user
RUN adduser -D -s /bin/sh quikgit

# Set working directory
WORKDIR /home/quikgit

# Copy the binary from GoReleaser
COPY quikgit /usr/local/bin/quikgit

# Make binary executable
RUN chmod +x /usr/local/bin/quikgit

# Switch to non-root user
USER quikgit

# Set entrypoint
ENTRYPOINT ["quikgit"]