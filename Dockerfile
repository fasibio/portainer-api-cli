FROM alpine:latest
RUN apk add ca-certificates
ADD ./portainer-api /bin/portainer-api