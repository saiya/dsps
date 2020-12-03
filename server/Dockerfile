FROM golang:1.15.5
ARG DSPS_VERSION_ID=""
WORKDIR /go/src/github.com/dsps/server
COPY . .
RUN go get && go get github.com/Songmu/gocredits/cmd/gocredits
RUN make dsps/dsps-$(uname -o | sed 's/GNU\///')-$(uname -m) && mv dsps/dsps-$(uname -o | sed 's/GNU\///')-$(uname -m) /dsps.bin

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /dsps.bin ./dsps
EXPOSE 3000
CMD ["./dsps"]
