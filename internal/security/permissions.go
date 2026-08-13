package security

// PermissionSet represents active runtime permissions.
type PermissionSet struct {
	AllowFileWrite  bool
	AllowFileDelete bool
	AllowShellExec  bool
	AllowNetwork    bool
}

// DefaultPermissions returns restricted baseline permissions.
func DefaultPermissions() PermissionSet {
	return PermissionSet{
		AllowFileWrite:  true,
		AllowFileDelete: false,
		AllowShellExec:  false,
		AllowNetwork:    false,
	}
}
