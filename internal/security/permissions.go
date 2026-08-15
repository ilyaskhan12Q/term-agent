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
		AllowShellExec:  true, // Enabled by default subject to POSIX AST security policy classification
		AllowNetwork:    false,
	}
}

// FullPermissions returns unrestricted development permissions.
func FullPermissions() PermissionSet {
	return PermissionSet{
		AllowFileWrite:  true,
		AllowFileDelete: true,
		AllowShellExec:  true,
		AllowNetwork:    true,
	}
}

// ReadOnlyPermissions returns strictly read-only permissions.
func ReadOnlyPermissions() PermissionSet {
	return PermissionSet{
		AllowFileWrite:  false,
		AllowFileDelete: false,
		AllowShellExec:  false,
		AllowNetwork:    false,
	}
}
