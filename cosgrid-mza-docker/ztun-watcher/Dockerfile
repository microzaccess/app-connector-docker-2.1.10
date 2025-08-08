FROM ubuntu:22.04

RUN apt-get update && apt-get install -y iproute2 sudo net-tools softflowd strace iputils-ping

COPY shared/opt/cosgrid /opt/cosgrid
COPY ztun-watcher/entrypoint.sh /entrypoint.sh
RUN chmod +x /opt/cosgrid/* /entrypoint.sh

CMD ["/entrypoint.sh"]
