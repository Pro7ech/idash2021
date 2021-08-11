FROM golang:1.16

RUN apt-get update; apt-get install time

COPY . /root/idash21_Track2

WORKDIR /root/idash21_Track2/prediction
RUN make build

ENTRYPOINT ["/bin/bash"]
