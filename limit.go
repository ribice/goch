package goch

// Limit represents limit type
type Limit int

// Limit constants
const (
	DisplayNameLimit Limit = iota + 1
	UIDLimit
	SecretLimit
	ChanLimit
	ChanSecretLimit
)
