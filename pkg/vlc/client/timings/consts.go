package timings

import "time"

const (
	StatusClarificationInterval            = 1000 * time.Millisecond
	WaitUntilOnlinePollingInterval         = 20 * time.Millisecond
	WaitForAutoSeekAfterFileOpenedDuration = 1000 * time.Millisecond
	CommandsRepeatInterval                 = 50 * time.Millisecond
	WaitForShutdownAfterStopDuration       = 500 * time.Millisecond

	SkipFollowerUpdatesBeforePollingIntervalsNumber = 1.5
)

func GetFollowerUpdatesIgnoreDuration(syncInterval time.Duration) time.Duration {
	return time.Duration(float64(syncInterval) * SkipFollowerUpdatesBeforePollingIntervalsNumber)
}
