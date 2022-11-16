FROM docker.io/library/golang:1.19 as builder

WORKDIR /kubesec

COPY cmd ./cmd
COPY vendor ./vendor
COPY go.mod .
COPY go.sum .
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kubesec-webhook ./cmd/kubesec

FROM docker.io/library/alpine:3.16

RUN addgroup -S kubesec \
  && adduser -S -g kubesec kubesec \
  && apk --no-cache add ca-certificates

WORKDIR /home/kubesec

COPY --from=builder /kubesec/kubesec-webhook .

RUN chown -R kubesec:kubesec ./

USER kubesec

CMD ["./kubesec-webhook"]
