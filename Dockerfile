# ================================================
# Build the binary using a temporary builder image
# ================================================

FROM golang:latest AS builder

# We'll need the source files.
WORKDIR /go/src/app
COPY . .

# Cross-compile the binary and strip as much as possible.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"

# ===========================================
# A much smaller image for the final artefact
# ===========================================

FROM scratch

COPY --from=builder /go/src/app/senko /bin/senko
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/bin/senko"]