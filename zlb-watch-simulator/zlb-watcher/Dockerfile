# Stage 1: build
FROM rust:1.73-bullseye AS builder

WORKDIR /usr/src/zlb-watcher
COPY . .

RUN cargo install --path .
#RUN cargo build --release

# Stage 2: runtime
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/src/zlb-watcher/target/release/zlb-watcher /usr/local/bin/zlb-watcher

ENTRYPOINT ["/usr/local/bin/zlb-watcher"]

