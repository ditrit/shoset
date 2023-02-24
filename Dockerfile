FROM ubuntu:latest as setup-build-stage-git
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y git
# temporary until merge to master
RUN git clone -b new_pki https://github.com/ditrit/shoset.git

FROM golang:1.18-bullseye as setup-build-stage-shoset
WORKDIR /app
COPY --from=setup-build-stage-git . /app
WORKDIR /app/shoset
RUN go get
WORKDIR /app/shoset/test
RUN go build -o shoset

FROM ubuntu:21.04 as setup-stage-production
COPY --from=setup-build-stage-shoset /app/shoset/test/shoset usr/bin/
RUN apt-get update

FROM setup-stage-production as stage-production
WORKDIR /usr/bin
CMD [ "./shoset", "4" ]
