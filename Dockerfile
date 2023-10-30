FROM ubuntu:20.04
WORKDIR /
COPY  bin/router .
COPY  bin/router-cni .
COPY  bin/router .

