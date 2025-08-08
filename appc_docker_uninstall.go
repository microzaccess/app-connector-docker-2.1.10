package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// List of images to delete (repository:tag format)
	images := []string{
		"microzaccess/data-base:saas",
		"microzaccess/log-transferer:2.1.10",
		"microzaccess/proxy:1.0",
		"microzaccess/mza-agent:2.1.10",
		"microzaccess/ztun-watcher:2.1.10",
		"microzaccess/logtrimmer:2.1.10",
	}

	fmt.Println("Starting Docker image cleanup...")

	for _, image := range images {
		fmt.Printf("Deleting image: %s\n", image)

		cmd := exec.Command("docker", "rmi", "-f", image)
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("Error deleting image %s: %v\n", image, err)
			fmt.Printf("Output: %s\n", string(output))
		} else {
			fmt.Printf("Successfully deleted image: %s\n", image)
		}
		fmt.Println(strings.Repeat("-", 50))
	}

	fmt.Println("Docker image cleanup completed!")

	// Remove the app-connector-docker-2.1.10 folder
	folderName := "app-connector-docker-2.1.10"
	fmt.Printf("\nRemoving folder: %s\n", folderName)

	err := os.RemoveAll(folderName)
	if err != nil {
		fmt.Printf("Error removing folder %s: %v\n", folderName, err)
	} else {
		fmt.Printf("Successfully removed folder: %s\n", folderName)
	}
}
