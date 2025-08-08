package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	baseDir := "app-connector-docker-2.1.10"

	// Create the main directory
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatalf("Failed to create base directory: %v", err)
	}

	// Create 'shared' directory inside baseDir
	if err := os.MkdirAll(baseDir+"/shared/.config/cosgrid", 0755); err != nil {
		log.Fatalf("Failed to create nested config directory: %v", err)
	}

	// Create empty files
	files := []string{
		"cosgrid-microzaccess.log",
		"mza_connecter_id.json",
		"service.log",
		".config/cosgrid/config.json",
		"acl_custom.log",
	}
	for _, file := range files {
		path := baseDir + "/shared/" + file
		if _, err := os.Create(path); err != nil {
			log.Fatalf("Failed to create file %s: %v", path, err)
		}
		if file == "acl_custom.log" {
			if err := os.Chmod(path, 0777); err != nil {
				log.Fatalf("Failed to chmod 777 %s: %v", path, err)
			}
		}
	}

	// Detect Docker Compose command
	method := "docker compose"
	cmd := exec.Command("docker-compose", "version")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("sudo", "docker-compose", "version")
		if err := cmd.Run(); err != nil {
			cmd = exec.Command("docker", "compose", "version")
			if err := cmd.Run(); err != nil {
				cmd = exec.Command("sudo", "docker", "compose", "version")
				if err := cmd.Run(); err != nil {
					log.Fatalf("docker-compose or docker compose command not found: %v", err)
				} else {
					method = "sudo docker compose"
				}
			} else {
				method = "docker compose"
			}
		} else {
			method = "sudo docker-compose"
		}
	} else {
		method = "docker-compose"
	}

	dockerContent := `
version: '3.8'
services:

  mza-agent:
    image: microzaccess/mza-agent:2.1.10
    container_name: mza-agent
    cap_add:
      - NET_ADMIN
    devices:
      - /dev/net/tun
    volumes:
      - ./shared/.config/cosgrid/config.json:/home/dockeruser/.config/cosgrid/config.json
      - /home/mza_connecter_id.json:/home/mza_connecter_id.json:ro
      - ./shared/service.log:/var/log/cosgrid/service.log
      - ./shared/cosgrid-microzaccess.log:/var/log/cosgrid/cosgrid-microzaccess.log 
    network_mode: host

  ztun-watcher:
    image: microzaccess/ztun-watcher:2.1.10
    container_name: ztun-watcher
    cap_add:
      - NET_ADMIN
      - NET_RAW
      - SYS_NICE  
      - DAC_OVERRIDE 
    devices:
      - /dev/net/tun
    network_mode: host
    pid: "container:mza-agent"
    depends_on:
    - mza-agent

  logtrimmer:
    image: microzaccess/logtrimmer:2.1.10
    container_name: logtrimmer
    volumes:
      - ./shared/service.log:/var/log/cosgrid/service.log
      - ./shared/cosgrid-microzaccess.log:/var/log/cosgrid/cosgrid-microzaccess.log 
    network_mode: host

  log-transferer:
    image: microzaccess/log-transferer:2.1.10
    container_name: log-transferer
    volumes:
      - ./shared/service.log:/var/log/cosgrid/service.log
      - ./shared/cosgrid-microzaccess.log:/var/log/cosgrid/cosgrid-microzaccess.log
      - ./shared/acl_custom.log:/var/log/cosgrid/acl_custom.log
    network_mode: host

  proxy:
    image: microzaccess/proxy:1.0
    container_name: proxy
    depends_on:
      - mza-agent
    volumes:
      - ./shared/acl_custom.log:/var/log/squid/acl_custom.log
    network_mode: host
    pid: "container:mza-agent"

  data-base:
    image: microzaccess/data-base:saas
    container_name: data-base
    volumes:
      - ./shared/.config/cosgrid/config.json:/home/.config/cosgrid/config.json
      - /home/mza_connecter_id.json:/home/mza_connecter_id.json:ro
      - /etc/ssl/certs:/etc/ssl/certs
    network_mode: host
    restart: unless-stopped
`

	dockerContent = strings.ReplaceAll(dockerContent, "\t", "    ")
	composeFile := baseDir + "/docker-compose.yml"
	if err := os.WriteFile(composeFile, []byte(dockerContent), 0644); err != nil {
		log.Fatalf("Failed to write docker-compose.yml: %v", err)
	}

	fmt.Println("Setup completed in", baseDir)
	fmt.Println("Use `cd", baseDir, "&&", method, "up -d` to start the containers.")
}
