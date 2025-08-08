#!/usr/bin/env python3

import sys
import json
import redis
import re
import logging
from urllib.parse import urlparse
import psutil

# Configure logging
logging.basicConfig(
    filename='/var/log/squid/acl_filter.log',
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)

extra_logger = logging.getLogger("extra_logger")
extra_logger.setLevel(logging.INFO)

extra_handler = logging.FileHandler('/var/log/squid/acl_custom.log')
extra_formatter = logging.Formatter('%(asctime)s - %(message)s')
extra_handler.setFormatter(extra_formatter)

extra_logger.addHandler(extra_handler)

# Redis connection
try:
    r = redis.Redis(host='127.0.0.1', port=6379, decode_responses=True)
    r.ping()  # Test connection
except Exception as e:
    logging.error(f"Failed to connect to Redis: {e}")
    sys.exit(1)

# Global allowlist domains
GLOBAL_ALLOWLIST = [
    "www.google.com", "accounts.google.com", "www.google.co.in", "cosgrid.net", "cosgridnetworks.in",
    "cosgrid.com", "vcdn.cosgrid.com", "maps.googleapis.com", "34.93.129.91", "api.geoapify.com",
    "cdn.jsdelivr.net", "cdnjs.cloudflare.com", "dpi.cosgrid.com", "firestore.googleapis.com",
    "fonts.googleapis.com", "fonts.gstatic.com", "maps.googleapis.com", "maps.gstatic.com",
    "mitre-attack.github.io", "rum.cosgrid.net", "stackpath.bootstrapcdn.com", "vcdn.cosgrid.com",
    "www.gstatic.com", "cdn.plot.ly", "code.jquery.com", "ids.cosgrid.com", "cosgridnetworks.com"
]

def get_ztun_source_ip():
    """Find the IP address of the interface starting with 'ztun-'"""
    interfaces = psutil.net_if_addrs()
    for iface, addrs in interfaces.items():
        if iface.startswith("ztun-"):
            for addr in addrs:
                if addr.family.name == 'AF_INET':  # IPv4
                    return addr.address
    return "unknown"

def normalize_domain(domain):
    """Normalize domain by removing trailing dots and local suffixes"""
    domain = domain.rstrip('.')
    domain = re.sub(r'\.lan$', '', domain)
    domain = re.sub(r'\.local$', '', domain)
    return domain.lower()

def extract_domain_from_url(url):
    """Extract domain from URL"""
    try:
        if not url.startswith(('http://', 'https://')):
            url = 'http://' + url
        parsed = urlparse(url)
        domain = parsed.netloc
        # Remove port if present
        if ':' in domain:
            domain = domain.split(':')[0]
        return normalize_domain(domain)
    except:
        return normalize_domain(url.split('/')[0])

def is_globally_allowed(domain):
    """Check if domain is in global allowlist"""
    for allowed_domain in GLOBAL_ALLOWLIST:
        if domain == allowed_domain or domain.endswith('.' + allowed_domain):
            logging.info(f"Domain {domain} is globally allowed")
            return True
    return False

def domain_matches(target_domain, rule_domain):
    """Check if target domain matches rule domain (supports wildcards)"""
    if rule_domain.startswith('*.'):
        # Wildcard match
        suffix = rule_domain[2:]  # Remove "*."
        return target_domain.endswith('.' + suffix) or target_domain == suffix
    else:
        # Exact match or subdomain match
        return target_domain == rule_domain or target_domain.endswith('.' + rule_domain)

def domain_matches_application(target_domain, rule_domain):
    """Check if target domain matches rule domain (supports wildcards)"""
    if rule_domain.startswith('*.'):
        # Wildcard match
        suffix = rule_domain[2:]  # Remove "*."
        return target_domain.endswith('.' + suffix) or target_domain == suffix
    else:
        # Exact match or subdomain match
        return target_domain == rule_domain 
    # or target_domain.endswith('.' + rule_domain)



def check_domain_access(ip, domain):
    """Main function to check if domain access should be allowed"""
    try:
        logging.info(f"Checking access for IP: {ip}, Domain: {domain}")
        
        # Check global allowlist first
        if is_globally_allowed(domain):
            return True
        
        # Get IP data from Redis
        ip_data_raw = r.get(f"ip:{ip}")
        if not ip_data_raw:
            logging.info(f"No IP data found for {ip}")
            return False  # Allow if no data found
        
        ip_data = json.loads(ip_data_raw)
        tenant_id = ip_data.get("tenant_id")
        user_id = ip_data.get("user_id")
        
        if not tenant_id or not user_id:
            logging.warning(f"Missing tenant_id or user_id for IP {ip}")
            return True
        
        logging.info(f"tenant_id: {tenant_id}, user_id: {user_id}")
        
        # Get domain policies from Redis
        domain_data_raw = r.get(f"tenant:{tenant_id}:user:{user_id}")
        if not domain_data_raw:
            logging.info(f"No domain data found for tenant:{tenant_id}:user:{user_id}")
            return True
        
        domain_data = json.loads(domain_data_raw)
        logging.info(f"Domain data loaded for user")
        
        # Extract policy lists
        whitelist_urls = domain_data.get("whitelist_urls", [])
        whitelist_applications = domain_data.get("whitelist_applications", [])
        whitelist_categories = domain_data.get("whitelist_categories", [])
        blacklist_urls = domain_data.get("blacklist_urls", [])
        blacklist_applications = domain_data.get("blacklist_applications", [])
        blacklist_categories = domain_data.get("blacklist_categories", [])
        rest_all = domain_data.get("rest_all", "allow")
        
        # Check whitelist URLs
        for allowed_domain in whitelist_urls:
            if domain_matches(domain, allowed_domain):
                logging.info(f"Domain {domain} found in whitelist_urls")
                return True
        
        # Check whitelist categories
        for allowed_domain in whitelist_categories:
            if domain_matches(domain, allowed_domain):
                logging.info(f"Domain {domain} found in whitelist_categories")
                return True
        
        # Check whitelist applications
        for allowed_domain in whitelist_applications:
            if domain_matches_application(domain, allowed_domain):
                logging.info(f"Domain {domain} found in whitelist_applications")
                return True
        
        # Check blacklist URLs
        for blocked_domain in blacklist_urls:
            if domain_matches(domain, blocked_domain):
                logging.info(f"Domain {domain} found in blacklist_urls")
                return False
        
        # Check blacklist categories
        for blocked_domain in blacklist_categories:
            if domain_matches(domain, blocked_domain):
                logging.info(f"Domain {domain} found in blacklist_categories")
                return False
        
        # Check blacklist applications
        for blocked_domain in blacklist_applications:
            if domain_matches_application(domain, blocked_domain):
                logging.info(f"Domain {domain} found in blacklist_applications")
                return False
        
        # Apply rest_all policy
        if rest_all == "deny":
            logging.info(f"Domain {domain} denied by rest_all policy")
            return False
        
        logging.info(f"Domain {domain} not allowed by default")
        return False
        
    except Exception as e:
        logging.error(f"Error processing request for IP {ip}, domain {domain}: {e}")
        return True  # Allow on error to avoid blocking all traffic

def grant():
    """Send OK response to Squid"""
    sys.stdout.write('OK\n')
    sys.stdout.flush()

def deny():
    """Send ERR response to Squid"""
    sys.stdout.write('ERR\n')
    sys.stdout.flush()

def main():
    """Main loop to process Squid requests"""
    logging.info("Squid ACL filter started")
    
    source_ip = get_ztun_source_ip()
    logging.info(f"Detected source interface IP: {source_ip}")

    while True:
        try:
            line = sys.stdin.readline().strip()
            if not line:
                continue
            
            # Parse input from Squid: IP URL
            parts = line.split()
            if len(parts) < 2:
                logging.warning(f"Invalid input format: {line}")
                deny()
                continue
            
            ip = parts[0]
            url = parts[1]
            logging.info(f"Received request for IP: {ip}, URL: {url}")
            # Extract domain from URL
            domain = extract_domain_from_url(url)
            
            # Check access
            if check_domain_access(ip, domain):
                extra_logger.info(f"source={source_ip} {ip} {url} ALLOWED")
                grant()
            else:
                extra_logger.info(f"source={source_ip} {ip} {url} DENIED")
                deny()
                
        except EOFError:
            break
        except Exception as e:
            logging.error(f"Unexpected error: {e}")
            deny()

if __name__ == "__main__":
    main()

