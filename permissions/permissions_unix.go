package permissions

import (
	"os"
	"os/user"
	"strconv"
)

// SetFilePermissions will set file permissions on the srcPath
func SetFilePermissions(srcPath string, userString string, groupString string, perms os.FileMode) (err error) {
	err = os.Chmod(srcPath, perms)
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
		err = os.Chown(srcPath, uid, gid)
		if err != nil {
			return err
		}
	}
	return
}
