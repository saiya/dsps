FROM golang:1.15.6
ARG DSPS_VERSION_ID=""
WORKDIR /go/src/github.com/dsps/server
COPY . .
RUN go get && go get github.com/Songmu/gocredits/cmd/gocredits
RUN make dsps/dsps-$(uname -o | sed 's/GNU\///')-$(uname -m) && mv dsps/dsps-$(uname -o | sed 's/GNU\///')-$(uname -m) /dsps.bin

FROM alpine:3.12.3  
# Add some utilities for convinience: gettext (contains envsubst), curl, jq
RUN apk --no-cache add ca-certificates gettext curl jq
WORKDIR /root/
COPY --from=0 /dsps.bin ./dsps
EXPOSE 3000
CMD ["./dsps"]
