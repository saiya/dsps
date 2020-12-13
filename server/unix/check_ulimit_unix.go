// +build linux darwin

package unix

import (
	"context"
	"syscall"

	"github.com/saiya/dsps/server/logger"
)

// NotifyUlimit generate warn logs if requirements not met.
func NotifyUlimit(ctx context.Context, desired UlimitRequirement) {
	var nofile syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &nofile); err != nil {
		logger.Of(ctx).Debugf(logger.CatStorage, "Could not get RLIMIT_NOFILE, skipped ulimit check: %v", err)
		return
	}
	if nofile.Cur < uint64(desired.NoFiles) {
		logger.Of(ctx).Warnf(logger.CatStorage, "Current value of RLIMIT_NOFILE is %d but current server configuration needs %d at least. Please consider to increase ulimit nofile.", nofile.Cur, desired.NoFiles)
	}
}
