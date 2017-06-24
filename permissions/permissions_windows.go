package permissions

import (
	"os"

	"github.com/hectane/go-acl"
	"github.com/hectane/go-acl/api"
)

// SetFilePermissions will set file permissions on the srcPath
func SetFilePermissions(srcPath string, userString string, groupString string, perms os.FileMode) (err error) {
	var accessList []api.ExplicitAccess

	if userString == "" {
		accessList = append(accessList, acl.GrantName((uint32(perms)&0700)<<23, "CREATOR OWNER"))
	} else {
		accessList = append(accessList, acl.GrantName((uint32(perms)&0700)<<23, userString))
	}

	if groupString == "" {
		accessList = append(accessList, acl.GrantName((uint32(perms)&0070)<<26, "CREATOR OWNER"))
	} else {
		accessList = append(accessList, acl.GrantName((uint32(perms)&0070)<<26, groupString))
	}

	return acl.Apply(
		srcPath,
		true,
		false,
		accessList...,
	)

	return
}
