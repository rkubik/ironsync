package main

import (
	"flag"
	"fmt"
	"ironsync/config"
	"ironsync/connection"
	"ironsync/permissions"
	"ironsync/resource"
	"ironsync/utils"
	"log"
	"os"
	"time"
)

// Command-line arguments
var (
	connFile = flag.String("connfile", "conn.ini", "Connection configuration file")
	resFile  = flag.String("resfile", "res.ini", "Resource configuration file")
)

// Program information
var (
	// ProgName - Program name
	progName = "ironsync"
	// ProgVersion - Current version
	progVersion = "0.1.0"
)

func processResource(c *connection.Connection, r *resource.Resource) (bool, error) {
	if r.PreUpdateCommand != "" {
		err := utils.RunCmd(r.PreUpdateCommand, r.PreUpdateCommandTimeout)
		if err != nil {
			return false, fmt.Errorf("Pre-update cmd failed: %v", err)
		}
	}

	modified, path, err := c.Download(r)
	if err != nil {
		return false, fmt.Errorf("Downloading resource failed: %v", err)
	}

	defer os.Remove(path)

	if !modified {
		return false, nil
	}

	// Avoid unnecessary overwrite if files are the same
	equal := utils.DeepCompare(path, r.Path)
	if equal {
		return false, nil
	}

	perms := r.Perms
	if perms == 0 {
		fileInfo, err := os.Stat(r.Path)
		if err != nil {
			perms = os.FileMode(int(0664))
		} else {
			perms = fileInfo.Mode()
		}
	}

	err = permissions.SetFilePermissions(path, r.User, r.Group, perms)
	if err != nil {
		return false, fmt.Errorf("Setting file permissions failed: %v", err)
	}

	err = os.Rename(path, r.Path)
	if err != nil {
		return false, fmt.Errorf("Moving file failed %s: %v", path, err)
	}

	if r.PostUpdateCommand != "" {
		err := utils.RunCmd(r.PostUpdateCommand, r.PostUpdateCommandTimeout)
		if err != nil {
			return false, fmt.Errorf("Post-update cmd failed: %v", err)
		}
	}

	return true, nil
}

func connectionWorker(c *connection.Connection) {
	log.Printf("[%s] Connected started", c.Name)

	for {
		for _, r := range c.Resources {
			if time.Now().After(r.NextUpdateTime) {
				log.Printf("[%s][%s] Updating resource", c.Name, r.Path)

				modified, err := processResource(c, r)
				if err != nil {
					log.Printf("[%s][%s] Resource failed to update: %v", c.Name, r.Path, err)
					r.SetNextUpdateTime(r.RetryInterval)
				} else {
					if modified {
						log.Printf("[%s][%s] Resource successfully updated", c.Name, r.Path)
						r.SetLastUpdateTime()
					} else {
						log.Printf("[%s][%s] Resource not modified", c.Name, r.Path)
					}
					r.SetNextUpdateTime(r.Interval)
				}
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func main() {
	flag.Parse()

	log.Printf("%s (version %s)", progName, progVersion)

	connections, err := config.Parse(*connFile, *resFile)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	} else if len(connections) == 0 {
		log.Fatalf("%s: No connections defined", *connFile)
	}

	for _, c := range connections {
		if len(c.Resources) > 0 {
			go connectionWorker(c)
		}
	}

	for {
		time.Sleep(1000 * time.Millisecond)
	}
}
