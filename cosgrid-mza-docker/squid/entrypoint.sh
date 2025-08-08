#!/bin/bash

# Ensure cache and log directories exist
mkdir -p /var/spool/squid
mkdir -p /var/log/squid



# Write squid.conf to /etc/squid/squid.conf
cat <<EOF > /etc/squid/squid.conf
# Visible hostname
visible_hostname RefleX-Edge

# Define local network access (adjust for your network)
acl localnet src 0.0.0.1-0.255.255.255
acl localnet src 10.0.0.0/8
acl localnet src 100.64.0.0/10
acl localnet src 169.254.0.0/16
acl localnet src 172.16.0.0/12
acl localnet src 192.168.0.0/16
#acl localnet src fc00::/7
#acl localnet src fe80::/10

# Define safe and SSL ports
acl SSL_ports port 443
acl Safe_ports port 80
acl Safe_ports port 21
acl Safe_ports port 443
acl Safe_ports port 70
acl Safe_ports port 210
acl Safe_ports port 1025-65535
acl Safe_ports port 280
acl Safe_ports port 488
acl Safe_ports port 591
acl Safe_ports port 777
acl CONNECT method CONNECT

# Logging format
logformat combined_format %>a %ssl::>sni %ru %>Hs %rm %un %tr %<st %mt
logfile_rotate 10
debug_options ALL,2 28,9

# Define listening ports for HTTP and HTTPS
http_port 31288
http_port 31289 intercept

#https_port 31290 cert=/etc/squid/ssl/ca.pem ssl-bump intercept
#https_port 31290 intercept ssl-bump cert=/etc/squid/ssl/squid.pem key=/etc/squid/ssl/squid.key 

# SSL Bump Configuration
acl step1 at_step SslBump1
acl step2 at_step SslBump2
acl step3 at_step SslBump3
ssl_bump peek step1 all
ssl_bump splice step2 all
ssl_bump bump step3 all

# Define network access rules
# acl mynetwork src 192.168.9.0/24
# http_access allow mynetwork

# acl localnet src all
# http_access allow localnet

external_acl_type domain_check_helper ttl=0 negative_ttl=0 %SRC %DST /etc/squid/handle_request.py
acl blocked_domains external domain_check_helper

http_access allow blocked_domains



# Whitelist URLs (both HTTP and HTTPS)
#acl whitelist url_regex -i chatgpt.com whatsapp.net whatsapp.com samsung.com jio.com www.google.com www.google.co.in stackpath.bootstrapcdn.com rum.cosgrid.net maps.gstatic.com vcdn.cosgrid.com leetcode.com cosgrid.com maps.googleapis.com api.geoapify.com cdn.jsdelivr.net www.gstatic.com accounts.google.com 34.93.129.91 dpi.cosgrid.com fonts.gstatic.com mitre-attack.github.io firestore.googleapis.com fonts.googleapis.com cosgrid.net cosgridnetworks.in cdnjs.cloudflare.com
# acl whitelist_https ssl::server_name_regex -i samsung.com www.google.com www.google.co.in stackpath.bootstrapcdn.com rum.cosgrid.net maps.gstatic.com vcdn.cosgrid.com leetcode.com cosgrid.com maps.googleapis.com api.geoapify.com cdn.jsdelivr.net www.gstatic.com accounts.google.com 34.93.129.91 dpi.cosgrid.com fonts.gstatic.com mitre-attack.github.io firestore.googleapis.com fonts.googleapis.com cosgrid.net cosgridnetworks.in cdnjs.cloudflare.com

#http_access allow whitelist
 #ssl_bump peek step1 whitelist_https
# ssl_bump splice step2 whitelist_https

# Allow only whitelisted HTTPS, everything else gets bumped
ssl_bump bump step3 all

# Allow localhost manager access
http_access allow localhost manager

# Deny all other access to the proxy
 http_access deny all

# Refresh Patterns
refresh_pattern ^ftp:        1440    20%    10080
refresh_pattern ^gopher:     1440    0%    1440
refresh_pattern -i (/cgi-bin/|\?) 0    0%    0
refresh_pattern .            0    20%    4320

# Caching and logging
cache_mem 256 MB
maximum_object_size_in_memory 8 KB
maximum_object_size 32 MB
cache_store_log /var/log/squid/cache.log

# SSL Certificate Storage
sslcrtd_program /usr/lib/squid/security_file_certgen -s /etc/squid/ssl_crtd -M 20MB

# Core dump directory
coredump_dir /var/cache/squid
EOF

# Run squid in foreground
exec squid -N -d1 -f /etc/squid/squid.conf
