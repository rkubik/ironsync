# IronSync

Project to help learn the Go Programming Language.

This service allows you to download file updates from many different remote
locations (HTTP, GitHub Gist, SFTP, FTP). Useful for syncing system settings,
key files, IDE preferences, password databases, etc.

## Building

    go build ironsync

## Running

    ./ironsync

or

    ./ironsync -connfile conn.ini -resfile res.ini

## Configuration

### Connections

Currently supported:

- HTTP
- GitHub Gist
- SFTP
- FTP
- Dropbox (OAuth 2)

Connection settings:

- `timeout`: Connection timeout (Default 30 sec)

HTTP settings:

- `url`: Complete URL (including protocol and port)
- `auth_username`: Basic HTTP Authentication
- `auth_password`: Basic HTTP Authentication

GitHub Gist settings:

- `url`: Custom URL (optional)
- `gist_username`: GitHub username
- `gist_id`: GitHub gist ID (32 character hex string)

SFTP settings:

- `hostname`: Hostname
- `port`: Port (Default 22)
- `auth_username`: Username
- `auth_password`: Password (optional)
- `private_key`: Private Key (optional)

FTP settings:

- `hostname`: Hostname
- `port`: Port (Default 22)
- `auth_username`: Username
- `auth_password`: Password

Dropbox settings:

- `dropbox_token`: OAuth 2 token

### Resources

Resource settings:

- `interval`: Number of seconds between successful updates (Default 60 sec)
- `retry_interval`: Number of seconds between failed updates (Default 30 sec)
- `user`: File user (optional)
- `group`: 'File group (optional)
- `perms`: File permissions (Default 0644)

HTTP settings:

- `remote_path`: Appended to connection URL (optional)

GitHub Gist settings:

- `remote_path`: Gist file, if Gist ID refers to multi-file (optional)
- `gist_id`: GitHub Gist ID (32 character hex identifier)
- `gist_username`: GistHub Gist Username

SFTP settings:

- `remote_path`: File path on SFTP server

FTP settings:

- `remote_path`: File path on FTP server

Dropbox settings:

- `remote_path`: Dropbox path format. See "Path formats"[1].

1. https://www.dropbox.com/developers/documentation/http/documentation

## TODO

- Persistent connections

## Examples

HTTP Example:

    [https]
    type = http
    url = https://myserver.com/file.txt

    [httpauth]
    type = http
    url = http://myserver2.com/file.txt
    auth_username = test
    auth_password = test

GitHub Gist Example:

    [github]
    type = gist

    [ghe]
    type = gist
    url = https://ghe.com/content

SFTP Example:

    [sftp]
    type = sftp
    hostname = myserver.com
    port = 2222
    auth_username = root
    private_key = /etc/myserver/id_rsa

FTP Example:

    [ftp]
    type = sftp
    hostname = myserver.com
    port = 2222
    auth_username = root
    auth_password = root

Dropbox Example:

    [dropbox]
    type = dropbox
    dropbox_token = k29e0fj49g82gh98gh24f49h
