package id

import (
	"time"

	"github.com/smedje/smedje/pkg/forge"
)

func optTime(opts forge.Options) time.Time {
	if opts.Time != nil {
		return opts.Time()
	}
	return time.Now()
}
