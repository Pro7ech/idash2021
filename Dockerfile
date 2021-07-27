FROM golang:1.12

COPY . /root/idash21_Track2

WORKDIR /root/idash21_Track2

ENTRYPOINT ["/bin/bash"]
