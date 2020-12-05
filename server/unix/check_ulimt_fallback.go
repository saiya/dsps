// +build !linux,!darwin

package unix

import (
	"context"
)

// NotifyUlimit generate warn logs if requirements not met.
func NotifyUlimit(ctx context.Context, desired UlimitRequirement) {}
