package resource

import (
	"os"
	"time"
)

// Resource - Filesystem resource
type Resource struct {
	Path       string // Absolute file path
	RemotePath string // ConnectionTypeHTTP: If set, appended to the URL, ConnectionTypeGist: Gist file (optional) */
	// Configuration
	Interval      int // Seconds
	RetryInterval int // Seconds
	// File attributes
	User  string      // User for UID
	Group string      // Group for GID
	Perms os.FileMode // File permissions
	// State
	NextUpdate time.Time // Time of next update
	LastUpdate time.Time // Time of last successful update
}

// CreateResource - Create a new resource object
func CreateResource(path string) Resource {
	return Resource{path, "", 60, 30, "", "", 0, time.Time{}, time.Time{}}
}

// SetNextUpdate - Set next update to given interval
func (r *Resource) SetNextUpdate(interval int) {
	r.NextUpdate = time.Now().Add(time.Second * time.Duration(interval))
}

// SetLastUpdate - Set last update to current time
func (r *Resource) SetLastUpdate() {
	r.LastUpdate = time.Now()
}
