package config

import (
	"fmt"
	"ironsync/connection"
	"ironsync/resource"
	"os"
	"strconv"

	"github.com/robfig/config"
)

func findConnection(name string, connections []*connection.Connection) (c *connection.Connection) {
	for _, c := range connections {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func parseConnectionConfig(connFile string) (connections []*connection.Connection, err error) {
	c, err := config.ReadDefault(connFile)
	if err != nil {
		return
	}

	for _, section := range c.Sections() {
		// Skip default section, unused
		if section == "DEFAULT" {
			continue
		}

		if findConnection(section, connections) != nil {
			return connections, fmt.Errorf("%s: Connection defined twice %s", connFile, section)
		}

		connType, err := c.String(section, "type")
		if err != nil {
			return connections, fmt.Errorf("%s: Section %s missing type", connFile, section)
		}

		if connType == "gist" {
			conn := connection.CreateGitHubGistConnection(section)

			// Optional
			connURL, err := c.String(section, "url")
			if err == nil {
				conn.URL = connURL
			}

			connTimeout, err := c.Int(section, "timeout")
			if err == nil {
				conn.Timeout = connTimeout
			}

			connections = append(connections, &conn)
		} else if connType == "http" {
			// Required
			connURL, err := c.String(section, "url")
			if err != nil {
				return connections, fmt.Errorf("%s: Section %s missing url", connFile, section)
			}

			conn := connection.CreateHTTPConnection(section, connURL)

			// Optional
			connTimeout, err := c.Int(section, "timeout")
			if err == nil {
				conn.Timeout = connTimeout
			}

			connections = append(connections, &conn)
		} else if connType == "sftp" {
			// Required
			connHostname, err := c.String(section, "hostname")
			if err != nil {
				return connections, fmt.Errorf("%s: Section %s missing hostname", connFile, section)

			}

			connAuthUsername, err := c.String(section, "auth_username")
			if err != nil {
				return connections, fmt.Errorf("%s: Section %s missing auth_username", connFile, section)

			}

			conn := connection.CreateSFTPConnection(section, connHostname, connAuthUsername)

			// Optional
			connTimeout, err := c.Int(section, "timeout")
			if err == nil {
				conn.Timeout = connTimeout
			}

			connAuthPassword, err := c.String(section, "auth_password")
			if err == nil {
				conn.AuthPassword = connAuthPassword
			}

			connPrivateKey, err := c.String(section, "private_key")
			if err == nil {
				conn.PrivateKey = connPrivateKey
			}

			connPort, err := c.Int(section, "port")
			if err == nil {
				conn.Port = connPort
			}

			connPersistent, err := c.Bool(section, "persistent")
			if err == nil {
				conn.Persistent = connPersistent
			}

			connections = append(connections, &conn)
		} else if connType == "ftp" {
			// Required
			connHostname, err := c.String(section, "hostname")
			if err != nil {
				return connections, fmt.Errorf("%s: Section %s missing hostname", connFile, section)

			}

			connAuthUsername, err := c.String(section, "auth_username")
			if err != nil {
				return connections, fmt.Errorf("%s: Section %s missing auth_username", connFile, section)
			}

			connAuthPassword, err := c.String(section, "auth_password")
			if err != nil {
				return connections, fmt.Errorf("%s: Section %s missing auth_password", connFile, section)
			}

			conn := connection.CreateFTPConnection(section, connHostname, connAuthUsername, connAuthPassword)

			// Optional
			connTimeout, err := c.Int(section, "timeout")
			if err == nil {
				conn.Timeout = connTimeout
			}

			connPort, err := c.Int(section, "port")
			if err == nil {
				conn.Port = connPort
			}

			connPersistent, err := c.Bool(section, "persistent")
			if err == nil {
				conn.Persistent = connPersistent
			}

			connections = append(connections, &conn)
		} else if connType == "dropbox" {
			// Required
			connDropboxToken, err := c.String(section, "dropbox_token")
			if err != nil {
				return connections, fmt.Errorf("%s: Section %s missing dropbox_token", connFile, section)
			}

			conn := connection.CreateDropboxConnection(section, connDropboxToken)

			connPersistent, err := c.Bool(section, "persistent")
			if err == nil {
				conn.Persistent = connPersistent
			}

			connections = append(connections, &conn)
		} else {
			return connections, fmt.Errorf("%s: Section %s invalid type %s", connFile, section, connType)
		}
	}

	return
}

func parseResourceConfig(resConfig string, connections []*connection.Connection) (err error) {
	c, err := config.ReadDefault(resConfig)
	if err != nil {
		return
	}

	for _, section := range c.Sections() {
		// Skip default section, unused
		if section == "DEFAULT" {
			continue
		}

		connName, err := c.String(section, "connection")
		if err != nil {
			return fmt.Errorf("%s: Section %s missing connection", resConfig, section)
		}

		conn := findConnection(connName, connections)
		if conn == nil {
			return fmt.Errorf("%s: Section %s invalid connection %s", resConfig, section, connName)
		}

		res := resource.CreateResource(section)

		resStat, err := os.Stat(section)
		if err == nil {
			res.LastModifiedTime = resStat.ModTime()
		}

		// Optional
		resInterval, err := c.Int(section, "interval")
		if err == nil {
			res.Interval = resInterval
		}

		resRetryInterval, err := c.Int(section, "retry_interval")
		if err == nil {
			res.RetryInterval = resRetryInterval
		}

		resUser, err := c.String(section, "user")
		if err == nil {
			res.User = resUser
		}

		resGroup, err := c.String(section, "group")
		if err == nil {
			res.Group = resGroup
		}

		resPerms, err := c.String(section, "perms")
		if err == nil {
			// String to octal
			resPermsInt, err := strconv.ParseInt(resPerms, 8, 64)
			if err != nil {
				return fmt.Errorf("%s: Section %s invalid perms %s", resConfig, section, resPerms)
			}
			res.Perms = os.FileMode(resPermsInt)
		}

		resPreUpdateCommand, err := c.String(section, "pre_update_cmd")
		if err == nil {
			res.PreUpdateCommand = resPreUpdateCommand
		}

		resPreUpdateCommandTimeout, err := c.Int(section, "pre_update_cmd_timeout")
		if err == nil {
			res.PreUpdateCommandTimeout = resPreUpdateCommandTimeout
		}

		resPostUpdateCommand, err := c.String(section, "post_update_cmd")
		if err == nil {
			res.PostUpdateCommand = resPostUpdateCommand
		}

		resPostUpdateCommandTimeout, err := c.Int(section, "post_update_cmd_timeout")
		if err == nil {
			res.PostUpdateCommandTimeout = resPostUpdateCommandTimeout
		}

		// Required (based on connection type)
		resRemotePath, err := c.String(section, "remote_path")
		if err == nil {
			res.RemotePath = resRemotePath
		} else if conn.Type == connection.ConnectionTypeFTP ||
			conn.Type == connection.ConnectionTypeSFTP ||
			conn.Type == connection.ConnectionTypeDropbox {
			return fmt.Errorf("%s: Section %s missing remote_path", resConfig, section)
		}

		if conn.Type == connection.ConnectionTypeGitHubGist {
			resGistID, err := c.String(section, "gist_id")
			if err != nil {
				return fmt.Errorf("%s: Section %s missing gist_id", resConfig, section)
			}
			res.GistID = resGistID

			resGitHubUsername, err := c.String(section, "github_username")
			if err != nil {
				return fmt.Errorf("%s: Section %s missing github_username", resConfig, section)
			}
			res.GitHubUsername = resGitHubUsername

			resGitHubToken, err := c.String(section, "github_token")
			if err == nil {
				res.GitHubToken = resGitHubToken
			}
		}

		conn.Resources = append(conn.Resources, &res)
	}

	return
}

// Parse - Parse connection and resource settings
func Parse(connFile string, resFile string) (connections []*connection.Connection, err error) {
	connections, err = parseConnectionConfig(connFile)
	if err != nil {
		return
	}
	err = parseResourceConfig(resFile, connections)
	return
}
