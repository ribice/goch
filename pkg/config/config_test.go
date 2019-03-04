package config_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/ribice/goch"

	"github.com/ribice/goch/pkg/config"

	"github.com/stretchr/testify/assert"
)

var (
	lims = map[goch.Limit][2]int{
		goch.DisplayNameLimit: [2]int{3, 128},
		goch.UIDLimit:         [2]int{20, 20},
		goch.SecretLimit:      [2]int{20, 50},
		goch.ChanLimit:        [2]int{10, 20},
		goch.ChanSecretLimit:  [2]int{20, 20},
	}
	limErrs = map[goch.Limit]error{
		goch.DisplayNameLimit: errors.New("displayName must be between 3 and 128 characters long"),
		goch.UIDLimit:         errors.New("uid must be between 20 and 20 characters long"),
		goch.SecretLimit:      errors.New("secret must be between 20 and 50 characters long"),
		goch.ChanLimit:        errors.New("channel must be between 10 and 20 characters long"),
		goch.ChanSecretLimit:  errors.New("channelSecret must be between 20 and 20 characters long"),
	}
)

func TestLoad(t *testing.T) {
	type data struct {
		user      string
		pass      string
		redisPass string
	}
	cases := []struct {
		name     string
		path     string
		wantData *config.Config
		wantErr  bool
		envData  *data
	}{
		{
			name:    "Fail on non-existing file",
			path:    "notExists",
			wantErr: true,
		},
		{
			name:    "Fail on wrong file format",
			path:    "testdata/invalid.yaml",
			wantErr: true,
		},
		{
			name:    "Fail on incorrect number of limits",
			path:    "testdata/limits.yaml",
			wantErr: true,
		},
		{
			name:    "Missing env vars",
			path:    "testdata/testdata.yaml",
			wantErr: true,
		},
		{
			name: "Missing pass env var",
			path: "testdata/testdata.yaml",
			envData: &data{
				user: "username",
			},
			wantErr: true,
		},
		{
			name: "Success",
			path: "testdata/testdata.yaml",
			wantData: &config.Config{
				Server: &config.Server{
					Port: 8080,
				},
				Redis: &config.Redis{
					Address:  "test.com",
					Port:     6379,
					Password: "repassword",
				},
				NATS: &config.NATS{
					ClusterID: "test-cluster",
					ClientID:  "test-client",
					URL:       "test-url",
				},
				Admin: &config.AdminAccount{
					Username: "admin",
					Password: "password",
				},
				Limits:    lims,
				LimitErrs: limErrs,
			},
			envData: &data{
				user:      "admin",
				pass:      "password",
				redisPass: "repassword",
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envData != nil {
				os.Setenv("REDIS_PASSWORD", tt.envData.redisPass)
				os.Setenv("ADMIN_USERNAME", tt.envData.user)
				os.Setenv("ADMIN_PASSWORD", tt.envData.pass)
			}
			cfg, err := config.Load(tt.path)
			fmt.Println(err)
			assert.Equal(t, tt.wantData, cfg)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestExceedsAny(t *testing.T) {
	cases := []struct {
		name    string
		req     map[string]goch.Limit
		wantErr error
	}{
		{
			name: "Exceeds DisplayName",
			req: map[string]goch.Limit{
				"TT": goch.DisplayNameLimit,
			},
			wantErr: errors.New("displayName must be between 3 and 128 characters long"),
		},
		{
			name: "Exceeds UID",
			req: map[string]goch.Limit{
				"TTT": goch.DisplayNameLimit,
				"UID": goch.UIDLimit,
			},
			wantErr: errors.New("uid must be between 20 and 20 characters long"),
		},
		{
			name: "Success",
			req: map[string]goch.Limit{
				"TTT":        goch.DisplayNameLimit,
				"1234567890": goch.ChanLimit,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Limits: map[goch.Limit][2]int{
					goch.DisplayNameLimit: [2]int{3, 128},
					goch.UIDLimit:         [2]int{20, 20},
					goch.ChanLimit:        [2]int{10, 20},
				},
				LimitErrs: map[goch.Limit]error{
					goch.DisplayNameLimit: errors.New("displayName must be between 3 and 128 characters long"),
					goch.UIDLimit:         errors.New("uid must be between 20 and 20 characters long"),
				},
			}
			assert.Equal(t, tt.wantErr, cfg.ExceedsAny(tt.req))
		})
	}
}

func TestExceeds(t *testing.T) {
	cases := []struct {
		name    string
		req     string
		limit   goch.Limit
		wantErr error
	}{
		{
			name:    "Fail",
			req:     "TT",
			limit:   goch.DisplayNameLimit,
			wantErr: errors.New("displayName must be between 3 and 128 characters long"),
		},
		{
			name:  "Success",
			req:   "TTT",
			limit: goch.DisplayNameLimit,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Limits: map[goch.Limit][2]int{
					goch.DisplayNameLimit: [2]int{3, 128},
				},
				LimitErrs: map[goch.Limit]error{
					goch.DisplayNameLimit: errors.New("displayName must be between 3 and 128 characters long"),
				},
			}
			assert.Equal(t, tt.wantErr, cfg.Exceeds(tt.req, tt.limit))
		})
	}
}
