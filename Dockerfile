FROM nixos/nix:latest AS builder

COPY . /tmp/build
WORKDIR /tmp/build

RUN nix \
    --extra-experimental-features "nix-command flakes" \
    --option filter-syscalls false \
    build

RUN mkdir /tmp/nix-store-closure
RUN cp -R $(nix-store -qR result/) /tmp/nix-store-closure

RUN mkdir -p /tmp/certs/etc/ssl/certs
RUN cp /etc/ssl/certs/ca-certificates.crt /tmp/certs/etc/ssl/certs/ca-certificates.crt

RUN mkdir -p /tmp/empty_tmp

FROM scratch

WORKDIR /app

COPY --from=builder /tmp/empty_tmp /tmp
COPY --from=builder /tmp/certs/etc/ssl/certs /etc/ssl/certs
COPY --from=builder /tmp/nix-store-closure /nix/store
COPY --from=builder /tmp/build/result /app

CMD ["/app/bin/short-form-discord-app"]
