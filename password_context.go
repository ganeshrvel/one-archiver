package onearchiver

type PasswordContext struct {
	passwords []string
}

// hasPasswords checks if there's at least one password in the context.
func (pd *PasswordContext) hasPasswords() bool {
	return len(pd.passwords) > 0
}

// getSinglePassword retrieves the first password if available, otherwise returns an empty string.
func (pd *PasswordContext) getSinglePassword() string {
	if !pd.hasPasswords() {
		return ""
	}
	return pd.passwords[0]
}

// hasSinglePassword checks if there's a single password available in the context.
func (pd *PasswordContext) hasSinglePassword() bool {
	return pd.getSinglePassword() != ""
}
