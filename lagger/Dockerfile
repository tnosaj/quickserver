FROM ubuntu:latest

WORKDIR /work

# install useful tools
RUN apt-get update && \
    apt-get install curl vim iproute2 iputils-ping dnsutils net-tools netcat-traditional socat -y

EXPOSE 8080
USER root

# copy files
WORKDIR /work
COPY startup.sh .
RUN chmod +x startup.sh

# keep running
ENTRYPOINT [ "/bin/bash", "/work/startup.sh" ]