FROM ubuntu:22.04

COPY shared/opt/cosgrid /opt/cosgrid
COPY logtrimmer/entrypoint.sh /entrypoint.sh
RUN chmod +x /opt/cosgrid/logtrimmer /entrypoint.sh

CMD ["/entrypoint.sh"]
