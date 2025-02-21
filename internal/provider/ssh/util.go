package ssh

import "strconv"

func ParsePermissions(perms string) uint32 {
	if perms == "" {
		return 0644
	}
	p, err := strconv.ParseUint(perms, 8, 32)
	if err != nil {
		return 0644
	}
	return uint32(p)
}
