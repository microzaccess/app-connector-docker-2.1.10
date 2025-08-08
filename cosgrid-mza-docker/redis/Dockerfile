# redis-service/Dockerfile

FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Install necessary packages
RUN apt-get update && \
    apt-get install -y redis-server iproute2 iptables inetutils-ping sudo net-tools && \
    rm -rf /var/lib/apt/lists/*


# Copy the static Go binary to the container
# COPY app_connector /home/app_connector

# Optional: Make sure it's executable
# RUN chmod +x /home/app_connector

# Optional: Custom Redis config (if needed)
# COPY redis.conf /etc/redis/redis.conf

# Create a startup script to run both services
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Start both app_connector and redis-server
CMD ["/usr/local/bin/entrypoint.sh"]
