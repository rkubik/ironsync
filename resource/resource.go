package resource

import (
	"os"
	"time"
)

const ()

// Resource - Filesystem resource
type Resource struct {
	Path       string // Absolute file path
	RemotePath string // ConnectionTypeHTTP: If set, appended to the URL, ConnectionTypeGist: Gist file (optional) */
	// Configuration
	Interval                 int    // Seconds
	RetryInterval            int    // Seconds
	PreUpdateCommand         string // Command to run before updating resource
	PreUpdateCommandTimeout  int    // Seconds
	PostUpdateCommand        string // Command to run after updating resource
	PostUpdateCommandTimeout int    // Seconds
	GistID                   string // GitHub Gist ID (32 character hex string)
	GitHubUsername           string // GitHub Username
	GitHubToken              string // GitHub OAuth2 Token
	// File attributes
	User  string      // User for UID
	Group string      // Group for GID
	Perms os.FileMode // File permissions
	// State
	NextUpdateTime   time.Time // Time of next update
	LastUpdateTime   time.Time // Time of last successful update (not accurate)
	LastModifiedTime time.Time // Last modified time on the server (accurate)
}

// CreateResource - Create a new resource object
func CreateResource(path string) Resource {
	return Resource{path, "", 60, 30, "", 10, "", 10, "", "", "", "", "", 0, time.Time{}, time.Time{}, time.Time{}}
}

// SetNextUpdateTime - Set next update to given interval
func (r *Resource) SetNextUpdateTime(interval int) {
	r.NextUpdateTime = time.Now().Add(time.Second * time.Duration(interval))
}

// SetLastUpdateTime - Set last update to current time
func (r *Resource) SetLastUpdateTime() {
	r.LastUpdateTime = time.Now()
}
