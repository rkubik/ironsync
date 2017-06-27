package connection

import (
	"fmt"
	"io"
	"io/ioutil"
	"ironsync/resource"
	"net"
	"net/http"
	"os"

	"time"

	"log"

	"ironsync/utils"

	"github.com/jlaffaye/ftp"
	"github.com/pkg/sftp"
	"github.com/tj/go-dropbox"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	// ConnectionTypeNone - Base connection type
	ConnectionTypeNone = 1
	// ConnectionTypeHTTP - HTTP connection
	ConnectionTypeHTTP = 2
	// ConnectionTypeGitHubGist - GitHub Gist (HTTP)
	ConnectionTypeGitHubGist = 3
	// ConnectionTypeSFTP SFTP connection
	ConnectionTypeSFTP = 4
	// ConnectionTypeFTP - FTP connection
	ConnectionTypeFTP = 5
	// ConnectionTypeDropbox - Dropbox connection
	ConnectionTypeDropbox = 6
)

const (
	// DefaultGitHubGistURL - Official GitGub Gist content URL
	DefaultGitHubGistURL = "https://gist.githubusercontent.com"
	// DefaultTimeout - Default connection timeout (seconds)
	DefaultTimeout = 30
	// DefaultMaxPacketSize - Default max packet size
	DefaultMaxPacketSize = 1 << 15
	// DefaultSSHPort - Default SSH port
	DefaultSSHPort = 22
	// DefaultFTPPort - Default FTP port
	DefaultFTPPort = 21
)

type downloadFunc func(*Connection, *resource.Resource, *os.File) (bool, error)

// Connection - Remote connection object
type Connection struct {
	// Data
	Name         string               // Unique connection name
	Type         int                  // Connection type
	Resources    []*resource.Resource // Array of Resources
	DownloadFunc downloadFunc         // Download function (nil if not set)

	// Connection objects
	SFTPClient *sftp.Client    // SFTP client (used for persistent connections)
	FTPClient  *ftp.ServerConn // FTP client (used for persistent connections)

	// Configuration
	Timeout       int  // Connection timouet (seconds)
	Persistent    bool // Keep a persistent connection
	URL           string
	Hostname      string
	Port          int
	MaxPacketSize int
	AuthUsername  string
	AuthPassword  string
	PrivateKey    string
	DropboxToken  string // Dropbox OAuth 2 access token
}

func downloadFTP(c *Connection, r *resource.Resource, tmpFile *os.File) (modified bool, err error) {
	if c.FTPClient == nil {
		addr := fmt.Sprintf("%s:%d", c.Hostname, c.Port)

		conn, err := ftp.DialTimeout(addr, time.Duration(c.Timeout)*time.Second)
		if err != nil {
			return modified, err
		}

		err = conn.Login(c.AuthUsername, c.AuthPassword)
		if err != nil {
			return modified, err
		}

		c.FTPClient = conn
	}

	if c.Persistent {
		defer c.FTPClient.Logout()
		defer c.FTPClient.Quit()
		defer func() {
			c.FTPClient = nil
		}()
	}

	remoteFile, err := c.FTPClient.Retr(r.RemotePath)
	if err != nil {
		return
	}
	defer remoteFile.Close()

	_, err = io.Copy(tmpFile, remoteFile)
	if err != nil {
		return
	}

	return true, err
}

func downloadSFTP(c *Connection, r *resource.Resource, tmpFile *os.File) (modified bool, err error) {
	if c.SFTPClient == nil {
		var auths []ssh.AuthMethod

		aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err == nil {
			auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
		}

		if c.AuthPassword != "" {
			auths = append(auths, ssh.Password(c.AuthPassword))
		}

		if c.PrivateKey != "" {
			auths = append(auths, utils.PublicKeyFile(c.PrivateKey))
		}

		sshConfig := ssh.ClientConfig{
			User:            c.AuthUsername,
			Auth:            auths,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Duration(c.Timeout) * time.Second,
		}

		addr := fmt.Sprintf("%s:%d", c.Hostname, c.Port)
		conn, err := ssh.Dial("tcp", addr, &sshConfig)
		if err != nil {
			return modified, err
		}

		sftpClient, err := sftp.NewClient(conn, sftp.MaxPacket(c.MaxPacketSize))
		if err != nil {
			conn.Close()
			return modified, err
		}
		c.SFTPClient = sftpClient
	}

	if !c.Persistent {
		defer c.SFTPClient.Close()
		defer func() {
			c.SFTPClient = nil
		}()
	}

	remoteFile, err := c.SFTPClient.Open(r.RemotePath)
	if err != nil {
		return
	}
	defer remoteFile.Close()

	_, err = io.Copy(tmpFile, remoteFile)
	if err != nil {
		return
	}

	return
}

func downloadHTTP(c *Connection, r *resource.Resource, tmpFile *os.File) (modified bool, err error) {
	url := c.URL

	if c.Type == ConnectionTypeGitHubGist {
		url = fmt.Sprintf("%s/%s/%s/raw/%s", url, r.GitHubUsername, r.GistID, r.RemotePath)
	} else if c.Type == ConnectionTypeHTTP && r.RemotePath != "" {
		url = fmt.Sprintf("%s/%s", url, r.RemotePath)
	}

	client := &http.Client{
		Timeout: time.Duration(c.Timeout) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	if c.Type == ConnectionTypeGitHubGist && r.GitHubToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", r.GitHubToken))
	}

	req.Header.Set("If-Modified-Since", r.LastModifiedTime.UTC().Format(http.TimeFormat))

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	lastModifiedStr := resp.Header.Get("Last-Modified")
	if lastModifiedStr != "" {
		lastModifiedTime, err := http.ParseTime(lastModifiedStr)
		if err == nil {
			if !lastModifiedTime.After(r.LastModifiedTime) {
				return false, err
			}
			r.LastModifiedTime = lastModifiedTime
		}
	}

	if resp.StatusCode == 304 {
		return false, err
	} else if resp.StatusCode != 200 {
		return false, fmt.Errorf("Connection failed to %s (%d)", url, resp.StatusCode)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	return true, err
}

func downloadDropbox(c *Connection, r *resource.Resource, tmpFile *os.File) (modified bool, err error) {
	config := dropbox.NewConfig(c.DropboxToken)
	config.HTTPClient.Timeout = time.Duration(c.Timeout) * time.Second

	db := dropbox.New(config)

	// Check ContentHash to see if file has been modifed. Continue on error.
	localHash, err := dropbox.FileContentHash(r.Path)
	if err == nil {
		metaData, err := db.Files.GetMetadata(&dropbox.GetMetadataInput{
			Path:             r.RemotePath,
			IncludeMediaInfo: false,
		})

		if err == nil && metaData.ContentHash == localHash {
			return false, nil
		}
	}

	remoteInput := dropbox.DownloadInput{Path: r.RemotePath}

	dstOutput, err := db.Files.Download(&remoteInput)
	if err != nil {
		return
	}

	_, err = io.Copy(tmpFile, dstOutput.Body)
	return true, err
}

// CreateConnection - Create a base connection
func CreateConnection(name string, connType int, connDownloadFunc downloadFunc) Connection {
	return Connection{name, connType, []*resource.Resource{}, connDownloadFunc, nil, nil, DefaultTimeout, false, "", "", 0, DefaultMaxPacketSize, "", "", "", ""}
}

// CreateHTTPConnection - Create a new HTTP connection
func CreateHTTPConnection(name string, url string) Connection {
	c := CreateConnection(name, ConnectionTypeHTTP, downloadHTTP)
	c.URL = url
	return c
}

// CreateGitHubGistConnection - Create a new GitHub gist connection
func CreateGitHubGistConnection(name string) Connection {
	c := CreateConnection(name, ConnectionTypeGitHubGist, downloadHTTP)
	c.URL = DefaultGitHubGistURL
	return c
}

// CreateSFTPConnection - Create a new SFTP connection
func CreateSFTPConnection(name, hostname, authUsername string) Connection {
	c := CreateConnection(name, ConnectionTypeSFTP, downloadSFTP)
	c.Port = DefaultSSHPort
	c.Hostname = hostname
	c.AuthUsername = authUsername
	return c
}

// CreateFTPConnection - Create a new FTP connection
func CreateFTPConnection(name, hostname, authUsername, authPassword string) Connection {
	c := CreateConnection(name, ConnectionTypeFTP, downloadFTP)
	c.Port = DefaultFTPPort
	c.Hostname = hostname
	c.AuthUsername = authUsername
	c.AuthPassword = authPassword
	return c
}

// CreateDropboxConnection - Create a new Dropbox connection
func CreateDropboxConnection(name, token string) Connection {
	c := CreateConnection(name, ConnectionTypeDropbox, downloadDropbox)
	c.DropboxToken = token
	return c
}

// Download - Download resource
func (c *Connection) Download(r *resource.Resource) (modified bool, path string, err error) {
	tmpFile, err := ioutil.TempFile("", c.Name)
	if err != nil {
		return
	}
	defer tmpFile.Close()

	if c.DownloadFunc == nil {
		log.Fatalf("Missing DownloadFunc for connection: %d", c.Type)
	}

	modified, err = c.DownloadFunc(c, r, tmpFile)
	if err != nil {
		defer os.Remove(tmpFile.Name())
	}
	return modified, tmpFile.Name(), err
}
