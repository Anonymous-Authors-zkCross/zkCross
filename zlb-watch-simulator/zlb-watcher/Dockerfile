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




# # Stage 1: Build the application
# FROM rust:1.78 as builder
# WORKDIR /usr/src/zlb-watcher
# COPY . .
# RUN cargo install --path .

# # Stage 2: Create the final, smaller image
# FROM debian:bullseye-slim
# RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
# COPY --from=builder /usr/src/zlb-watcher/target/release/zlb-watcher /usr/local/bin/zlb-watcher
# ENTRYPOINT ["/usr/local/bin/zlb-watcher"]
# CMD ["zlb-watcher"]


#FROM debian:bullseye-slim
#RUN apt-get update && apt-get install -y libssl-dev ca-certificates && rm -rf /var/lib/apt/lists/*
#COPY --from=builder /usr/local/cargo/bin/zlb-watcher /usr/local/bin/zlb-watcher
#CMD ["zlb-watcher"]