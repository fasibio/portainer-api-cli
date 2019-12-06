FROM golang:alpine as build_go_env
ADD . $GOPATH/src/fasibio/portainer-api-cli
RUN cd $GOPATH/src/fasibio/portainer-api-cli && GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -mod vendor -o /src/bin/portainer-api .

FROM alpine:3.9
COPY --from=build_go_env /src/bin/portainer-api /bin