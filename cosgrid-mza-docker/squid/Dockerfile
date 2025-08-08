FROM ubuntu:22.04

RUN apt-get update && \
    apt-get install -y squid-openssl python3 python3-pip && \
    rm -rf /var/lib/apt/lists/*

RUN pip3 install redis
RUN pip3 install psutil

COPY handle_request.py /etc/squid/handle_request.py
RUN chmod +x /etc/squid/handle_request.py
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
