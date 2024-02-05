package repetition

import (
	"time"

	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
)

type Rule struct {
	Interval typeutil.Optional[time.Duration]
}

func Single() Rule {
	return Rule{}
}

func WithInterval(interval time.Duration) Rule {
	return Rule{
		Interval: typeutil.NewOptional(interval),
	}
}
