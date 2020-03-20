# ================================================
# Build the binary using a temporary builder image
# ================================================

FROM golang:latest AS builder

# We'll need the source files.
WORKDIR /go/src/app
COPY . .

# Cross-compile the binary and strip as much as possible.
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"

# ===========================================
# A much smaller image for the final artefact
# ===========================================

FROM golang:latest

COPY --from=builder /go/src/app/senko /bin/senko

CMD ["senko"]