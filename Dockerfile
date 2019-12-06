FROM alpine:3.9
RUN apk add ca-certificates
ADD ./portainer-api /bin/portainer-api