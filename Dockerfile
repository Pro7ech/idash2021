FROM golang:1.16

COPY . /root/idash21_Track2

WORKDIR /root/idash21_Track2

ENTRYPOINT ["/bin/bash"]
