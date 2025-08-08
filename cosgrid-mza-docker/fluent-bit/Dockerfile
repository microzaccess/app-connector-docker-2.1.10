FROM ubuntu:22.04

RUN apt-get update && apt-get install -y sudo net-tools curl unzip gnupg2

COPY shared/setup_fluent_bit /setup_fluent_bit
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /setup_fluent_bit /entrypoint.sh

RUN /setup_fluent_bit

CMD ["/entrypoint.sh"]
