FROM golang:1.7.4

WORKDIR $GOPATH/src/github.com/ONSdigital/dp-ci/commit-verification/

COPY . .

RUN go build

ENTRYPOINT ./commit-verification
