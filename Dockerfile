FROM golang:1.19-alpine AS builder
COPY . /build
WORKDIR /build
RUN apk add --no-cache build-base && \
    go test -race -v ./... && \
    GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -tags netgo -a -v -o /build/kubesec-webhook /build/cmd/kubesec

FROM alpine:3.17.0

ENV USER=webhook
ENV GROUP=webhook
ENV HOMEDIR="/app"
ENV UID=60000
ENV GID=60000
ENV SHELL='/sbin/nologin'

#Usage: adduser [OPTIONS] USER [GROUP]
#
#Create new user, or add USER to GROUP
#
#	-h DIR		Home directory
#	-g GECOS	GECOS field
#	-s SHELL	Login shell
#	-G GRP		Group
#	-S		    Create a system user
#	-D		    Don't assign a password
#	-H		    Don't create home directory
#	-u UID		User id
#	-k SKEL		Skeleton directory (/etc/skel)

RUN addgroup  -g "${GID}" "${GROUP}" && \
    adduser  \
    -G ${GROUP}  \
    -D \
    -g '' \
    -h ${HOMEDIR} \
    -s "${SHELL}" \
    -u "${UID}" \
    "${USER}"

COPY --from=builder /build/kubesec-webhook /app/kubesec
USER "${USER}"
WORKDIR "${HOMEDIR}"

ENTRYPOINT [ "/app/kubesec" ]

