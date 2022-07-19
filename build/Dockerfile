FROM golang:1.18.1 as builder
ARG VERSION
ARG GIT_COMMIT
ARG DATE

WORKDIR /go/src/github.com/nginxinc/nginx-kubernetes-gateway/cmd/gateway

COPY go.mod go.sum /go/src/github.com/nginxinc/nginx-kubernetes-gateway
RUN go mod download

COPY cmd /go/src/github.com/nginxinc/nginx-kubernetes-gateway/cmd
COPY internal /go/src/github.com/nginxinc/nginx-kubernetes-gateway/internal
COPY pkg /go/src/github.com/nginxinc/nginx-kubernetes-gateway/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -a -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${GIT_COMMIT} -X main.date=${DATE}" -o gateway .

FROM alpine:3.16.1 as capabilizer
RUN apk add --no-cache libcap

FROM capabilizer as local-capabilizer
COPY ./build/.out/gateway /usr/bin/
RUN setcap 'cap_kill=+ep' /usr/bin/gateway

FROM capabilizer as container-capabilizer
COPY --from=builder /go/src/github.com/nginxinc/nginx-kubernetes-gateway/cmd/gateway/gateway /usr/bin/
RUN setcap 'cap_kill=+ep' /usr/bin/gateway

FROM scratch as common
USER 1001:1001
ENTRYPOINT [ "/usr/bin/gateway" ]

FROM common as container
COPY --from=container-capabilizer /usr/bin/gateway /usr/bin/

FROM common as local
COPY --from=local-capabilizer /usr/bin/gateway /usr/bin/
