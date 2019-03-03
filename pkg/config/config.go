package config

import "github.com/ribice/goch"

// Config represents application configuration
type Config struct {
	limits    map[goch.Limit][2]int
	limitErrs map[goch.Limit]error
}

// Exceeds checks whether a string exceeds chat limitation
func (c *Config) Exceeds(str string, lim goch.Limit) error {
	if exceedsLim(str, c.limits[lim]) {
		return c.limitErrs[lim]
	}
	return nil
}

func exceedsLim(s string, lims [2]int) bool {
	return len(s) < lims[0] || len(s) > lims[1]
}
