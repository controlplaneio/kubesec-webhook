FROM golang:1.10 as builder

RUN mkdir -p /go/src/github.com/stefanprodan/kubesec-webhook/

WORKDIR /go/src/github.com/stefanprodan/kubesec-webhook

COPY . .

#RUN go test $(go list ./... | grep -v integration | grep -v /vendor/ | grep -v /template/) -cover

RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") && \
  GIT_COMMIT=$(git rev-list -1 HEAD) && \
  CGO_ENABLED=0 GOOS=linux go build \
  -a -installsuffix cgo -o kubesec ./cmd/kubesec

FROM alpine:3.7

RUN addgroup -S app \
    && adduser -S -g app app \
    && apk --no-cache add \
    ca-certificates

WORKDIR /home/app

COPY --from=builder /go/src/github.com/stefanprodan/kubesec-webhook/kubesec .

RUN chown -R app:app ./

USER app

CMD ["./kubesec"]
