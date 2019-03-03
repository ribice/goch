package goch

// Limit represents limit type
type Limit int

// Limit constants
const (
	DisplayNameLimit Limit = iota
	UIDLimit
	SecretLimit
	ChanLimit
	ChanSecretLimit
)
