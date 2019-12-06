FROM alpine:3.9
COPY --from=build_go_env /src/bin/portainer-api /bin