package main

import (
	"flag"
	"ironsync/config"
	"ironsync/connection"
	"ironsync/permissions"
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

func connectionWorker(c *connection.Connection) {
	log.Printf("[%s] Connected started", c.Name)

	for {
		for _, r := range c.Resources {
			if time.Now().After(r.NextUpdateTime) {
				log.Printf("[%s][%s] Updating resource", c.Name, r.Path)

				modified, path, err := c.Download(r)
				if err != nil {
					log.Printf("[%s][%s] Downloading resource failed: %v", c.Name, r.Path, err)
					r.SetNextUpdateTime(r.RetryInterval)
					continue
				}

				if !modified {
					log.Printf("[%s][%s] Resource not modified; download not performed", c.Name, r.Path)
					r.SetNextUpdateTime(r.Interval)
					continue
				}

				// Avoid unnecessary overwrite if files are the same
				equal := utils.DeepCompare(path, r.Path)
				if equal {
					log.Printf("[%s][%s] Resource not modified", c.Name, r.Path)
					r.SetNextUpdateTime(r.Interval)
					continue
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
					log.Printf("[%s][%s] Setting file permissions failed: %v", c.Name, r.Path, err)
					r.SetNextUpdateTime(r.RetryInterval)
					continue
				}

				err = os.Rename(path, r.Path)
				if err != nil {
					log.Printf("[%s][%s] Moving file failed %s: %v", c.Name, r.Path, path, err)
					r.SetNextUpdateTime(r.RetryInterval)
					continue
				}

				log.Printf("[%s][%s] Resource successfully updated", c.Name, r.Path)

				r.SetNextUpdateTime(r.Interval)
				r.SetLastUpdateTime()
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
