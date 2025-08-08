FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    iproute2 iptables inetutils-ping sudo net-tools ca-certificates && \
    useradd -m dockeruser

COPY shared/opt/cosgrid /opt/cosgrid
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /opt/cosgrid/* /entrypoint.sh

CMD ["/entrypoint.sh"]
