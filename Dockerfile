ARG GO_VERSION

FROM golang:${GO_VERSION}

WORKDIR /build

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build

FROM scratch

USER 1000

COPY --from=0 /build/kube-better-node /bin/kube-better-node

ENTRYPOINT ["/bin/kube-better-node"]
