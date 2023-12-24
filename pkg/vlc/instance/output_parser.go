package instance

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"sync"
	"time"

	timeutil "github.com/cardinalby/vlc-sync-play/pkg/util/time"
)

type StdErrEvent int

const (
	StderrEventMouse1Click StdErrEvent = iota
	eventsCount
)

type EventsToParse [eventsCount]bool

func (e EventsToParse) HasMouse1Click() bool {
	return e[StderrEventMouse1Click]
}

func (e EventsToParse) SetMouse1Click(value bool) {
	e[StderrEventMouse1Click] = value
}

func (e EventsToParse) HasAny() bool {
	for _, v := range e {
		if v {
			return true
		}
	}
	return false
}

// [000000010f420160] message
const lineMessagePrefixSize = 19
const waitForMouseUpTimeout = time.Second

type internalEvent int

const (
	noInternalEvent internalEvent = iota
	mouse1ButtonDown
	mouse1ButtonUp
)

// OutputParser receives stdout data from VLC, separates it by lines (ending with "\n"), parses each line
// and calls onEvent callback for each parsed event.
type OutputParser struct {
	ctx               context.Context
	scanner           *bufio.Scanner
	eventsToParse     EventsToParse
	eventsToParseMu   sync.RWMutex
	mouse1downCounter *timeutil.TtlCounter
}

func NewOutputParser(
	pipe io.ReadCloser,
	eventsToParse EventsToParse,
) *OutputParser {
	return &OutputParser{
		scanner:           bufio.NewScanner(pipe),
		eventsToParse:     eventsToParse,
		mouse1downCounter: timeutil.NewTtlCounter(waitForMouseUpTimeout),
	}
}

func (p *OutputParser) SetEventsToParse(events EventsToParse) {
	p.eventsToParseMu.Lock()
	p.eventsToParse = events
	p.eventsToParseMu.Unlock()
}

func (p *OutputParser) Start(ctx context.Context, onEvent func(event StdErrEvent)) error {
	for p.scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if p.scanner.Err() != nil {
			return p.scanner.Err()
		}
		p.eventsToParseMu.RLock()
		if !p.eventsToParse.HasAny() {
			p.eventsToParseMu.RUnlock()
			continue
		}
		var eventsToParse EventsToParse
		copy(eventsToParse[:], p.eventsToParse[:])
		p.eventsToParseMu.RUnlock()

		if event, ok := p.parseLine(p.scanner.Bytes(), eventsToParse); ok {
			onEvent(event)
		}
	}
	return nil
}

func (p *OutputParser) parseLine(line []byte, eventsToParse EventsToParse) (StdErrEvent, bool) {
	intEvent := p.getInternalEvent(line, eventsToParse)
	switch intEvent {
	case mouse1ButtonDown:
		p.mouse1downCounter.Inc()
	case mouse1ButtonUp:
		if p.mouse1downCounter.TryDec() {
			return StderrEventMouse1Click, true
		}
	case noInternalEvent:
	}
	return 0, false
}

var mouse1ButtonMessagePrefix = []byte("main vout display debug: VoutDisplayEvent 'mouse button' 0 ")
var mouse1ButtonDownSuffix = []byte("t=6")
var mouse1ButtonUpSuffix = []byte("t=7")

func (p *OutputParser) getInternalEvent(line []byte, eventsToParse EventsToParse) internalEvent {
	if len(line) < lineMessagePrefixSize {
		return noInternalEvent
	}
	message := line[lineMessagePrefixSize:]

	if eventsToParse.HasMouse1Click() && bytes.HasPrefix(message, mouse1ButtonMessagePrefix) {
		if bytes.HasSuffix(message, mouse1ButtonDownSuffix) {
			return mouse1ButtonDown
		}
		if bytes.HasSuffix(message, mouse1ButtonUpSuffix) {
			return mouse1ButtonUp
		}
	}

	return noInternalEvent
}
