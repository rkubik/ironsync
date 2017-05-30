package main

import (
	"flag"
	"ironsync/config"
	"ironsync/connection"
	"ironsync/utils"
	"log"
	"os"
	"os/user"
	"strconv"
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

func setFilePermissions(fileSrc string, fileDst string, userString string, groupString string, perms os.FileMode) (err error) {
	// Permissions - If none are provided then use permissions of existing
	// file. If existing file does not exist then use default.
	fileMode := perms
	if fileMode == 0 {
		fileInfo, err := os.Stat(fileDst)
		if err != nil {
			if os.IsNotExist(err) {
				fileMode = os.FileMode(int(0664))
			} else {
				return err
			}
		}
		fileMode = fileInfo.Mode()
	}

	err = os.Chmod(fileSrc, fileMode)
	if err != nil {
		return
	}

	// User
	uid := -1
	if userString != "" {
		u, err := user.Lookup(userString)
		if err != nil {
			return err
		}

		uid, err = strconv.Atoi(u.Uid)
		if err != nil {
			return err
		}
	}

	// Group
	gid := -1
	if groupString != "" {
		g, err := user.LookupGroup(groupString)
		if err != nil {
			return err
		}

		gid, err = strconv.Atoi(g.Gid)
		if err != nil {
			return err
		}
	}

	if uid != -1 || gid != -1 {
		err = os.Chown(fileSrc, uid, gid)
		if err != nil {
			return err
		}
	}
	return
}

func connectionWorker(c *connection.Connection) {
	log.Printf("[%s] Connected started", c.Name)

	for {
		for _, r := range c.Resources {
			if time.Now().After(r.NextUpdate) {
				log.Printf("[%s][%s] Updating resource", c.Name, r.Path)

				modified, path, err := c.Download(r)
				if err != nil {
					log.Printf("[%s][%s] Downloading resource failed: %v", c.Name, r.Path, err)
					r.SetNextUpdate(r.RetryInterval)
					continue
				}

				if !modified {
					log.Printf("[%s][%s] Resource not modified; download not performed", c.Name, r.Path)
					r.SetNextUpdate(r.Interval)
					continue
				}

				// Avoid unnecessary overwrite if files are the same
				equal := utils.DeepCompare(path, r.Path)
				if equal {
					log.Printf("[%s][%s] Resource not modified", c.Name, r.Path)
					r.SetNextUpdate(r.Interval)
					continue
				}

				err = setFilePermissions(path, r.Path, r.User, r.Group, r.Perms)
				if err != nil {
					log.Printf("[%s][%s] Setting file permissions failed: %v", c.Name, r.Path, err)
					r.SetNextUpdate(r.RetryInterval)
					continue
				}

				err = os.Rename(path, r.Path)
				if err != nil {
					log.Printf("[%s][%s] Moving file failed %s: %v", c.Name, r.Path, path, err)
					r.SetNextUpdate(r.RetryInterval)
					continue
				}

				log.Printf("[%s][%s] Resource successfully updated", c.Name, r.Path)

				r.SetNextUpdate(r.Interval)
				r.SetLastUpdate()
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
