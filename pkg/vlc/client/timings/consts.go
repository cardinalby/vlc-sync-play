package timings

import "time"

const (
	StatusClarificationInterval            = 20 * time.Millisecond
	WaitUntilOnlinePollingInterval         = StatusClarificationInterval
	WaitForAutoSeekAfterFileOpenedDuration = 2000 * time.Millisecond
	CommandsRepeatInterval                 = 50 * time.Millisecond
)
