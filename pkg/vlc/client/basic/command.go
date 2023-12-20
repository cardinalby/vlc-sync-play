package basic

import (
	"fmt"
)

type Key string

const (
	KeyCommand Key = "command"
	KeyInput   Key = "input"
	KeyVal     Key = "val"
)

// Command is a Command to send to the VLC API.
// See: https://github.com/videolan/vlc/tree/master/share/lua/http/requests
type Command map[Key]string

func PauseCmd() Command {
	return Command{
		KeyCommand: "pl_forcepause",
	}
}

func ResumeCmd() Command {
	return Command{
		KeyCommand: "pl_forceresume",
	}
}

func StopCmd() Command {
	return Command{
		KeyCommand: "pl_stop",
	}
}

func SeekCmd(position float64) Command {
	return Command{
		KeyCommand: "seek",
		KeyVal:     fmt.Sprintf("%f", position*100) + "%",
	}
}

func RateCmd(rate float64) Command {
	return Command{
		KeyCommand: "rate",
		KeyVal:     fmt.Sprintf("%f", rate),
	}
}

func PlayFileCmd(input string) Command {
	return Command{
		KeyCommand: "in_play",
		KeyInput:   input,
	}
}
