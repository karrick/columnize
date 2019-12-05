package gologs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// DefaultCommandFormat specifies a log format might be more appropriate for a
// infrequently used command line program, where the name of the service is a
// recommended part of the log line, but the timestamp is not.
const DefaultCommandFormat = "{program}: {message}"

// DefaultServiceFormat specifies a log format might be more appropriate for a
// service daemon, where the name of the service is implied by the filename the
// logs will eventually be written to. The default timestamp format is the same
// as what the standard library logs times as, but different timestamp formats
// are readily available, and the timestamp format is also customizable.
const DefaultServiceFormat = "{timestamp} [{level}] {message}"

// Level type defines one of several possible log levels.
type Level uint32

const (
	// Debug is for events that might help a person understand the cause of a
	// bug in a program.
	Debug Level = iota

	// Verbose is for events that might help a person understand the state of a
	// program.
	Verbose

	// Info is for events that annotate high level status of a program.
	Info

	// Warning is for events that indicate a possible problem with the
	// program. Warning events should be investigated and corrected soon.
	Warning

	// Error is for events that indicate a definite problem that might prevent
	// normal program execution. Error events should be corrected immediately.
	Error
)

func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Verbose:
		return "VERBOSE"
	case Info:
		return "INFO"
	case Warning:
		return "WARNING"
	case Error:
		return "ERROR"
	}
	// NOT REACHED
	panic(fmt.Sprintf("invalid log level: %d", uint32(l)))
}

// event instances are created by loggers and flow through the log tree from the
// branch where they were created, down to the base, at which point, its
// arguments will be formatted immediately prior to writing the log message to
// the underlying log io.Writer.
type event struct {
	args   []interface{}
	prefix []string
	when   time.Time
	format string
	level  Level
	tracer bool
}

// base is at the bottom of the logger tree, and formats the event to a byte
// slice, ensuring it ends with a newline, and writes its output to its
// underlying io.Writer.
type base struct {
	formatters     []func(*event, *[]byte)
	w              io.Writer
	c              int // c is count of bytes to allocate for formatting log line
	m              sync.Mutex
	isTimeRequired bool
}

func (b *base) log(e *event) error {
	// ??? *If* want to sacrifice a bit of speed, might consider using a
	// pre-allocated byte slice to format the output. The pre-allocated slice
	// can be protected with the lock already being used to serialize output, or
	// even better, its own lock so one thread can be formatting an event while
	// a different thread is writing the formatted event to the underlying
	// writer.
	buf := make([]byte, 0, b.c)

	// NOTE: This logic allows for a race between two threads that both get the
	// time for an event, then race for the mutex below that serializes output
	// to the underlying io.Writer. While not dangerous, the logic might allow
	// two log lines to be emitted to the writer in opposite timestamp order.
	if b.isTimeRequired {
		e.when = time.Now()
	}

	// Format the event according to the compiled formatting functions created
	// when the logger was created, according to the log template, i.e.,
	// "{timestamp} [{level}] {message}".
	for _, formatter := range b.formatters {
		formatter(e, &buf)
	}
	buf = singleNewline(buf)

	// Serialize access to the underlying io.Writer.
	b.m.Lock()
	_, err := b.w.Write(buf)
	b.m.Unlock()
	return err
}

func singleNewline(buf []byte) []byte {
	l := len(buf)
	if l == 0 {
		return []byte{'\n'}
	}

	// While this is O(length s), it stops as soon as it finds the first non
	// newline character in the string starting from the right hand side of the
	// input string. Generally this only scans one or two characters and
	// returns.
	for i := l - 1; i >= 0; i-- {
		if buf[i] != '\n' {
			if i+1 < l && buf[i+1] == '\n' {
				return buf[:i+2]
			}
			return append(buf[:i+1], '\n')
		}
	}

	return buf[:1] // all newline characters, so just return the first one
}

type logger interface {
	log(*event) error
}

// Logger provides methods to create events to be logged. Logger instances are
// created to emit events to their parent Logger instance, which may themselves
// either filter events based on a configured level, or prefix events with a
// configured string.
//
// When a logger is in Error mode, only Error events are logged. When a logger
// is in Warning mode, only Error and Warning events are logged. When a logger
// is in Info mode, only Error, Warning, and Info events are logged. When a
// logger is in Verbose mode, only Error, Warning, Info, and Verbose events are
// logged. When a logger is in Debug mode, all events are logged.
type Logger struct {
	prefix string // prefix is an option string, that when not empty, will prefix events
	parent logger // parent is the logger this branch sends events to
	level  Level  // level is the independent log level controls for this branch
	tracer bool   // tracer is value used to initialize new events created by this Logger
}

// New returns a new Logger instance that emits logged events to w after
// formatting the event according to template.
//
// Logger instances returned by this function are initialized to Warning level,
// which I feel is in keeping with the UNIX philosophy to _Avoid unnecessary
// output_. Simple command line programs will not need to set the log level to
// prevent spewing too many log events. While service application developers are
// more likely to spend a few minutes to build in the ability to configure the
// log level based on their service needs.
func New(w io.Writer, template string) (*Logger, error) {
	formatters, isTimeRequired, err := compileFormat(template)
	if err != nil {
		return nil, err
	}
	// Create a dummy event to see how long the log line is with the provided
	// template.
	buf := make([]byte, 0, 64)
	var e event
	for _, formatter := range formatters {
		formatter(&e, &buf)
	}
	min := len(buf) + 64
	if min < 128 {
		min = 128
	}
	parent := &base{
		c:              min,
		formatters:     formatters,
		isTimeRequired: isTimeRequired,
		w:              w,
	}
	return &Logger{parent: parent, level: Warning}, nil
}

// NewBranch returns a new Logger instance that logs to parent, but has its own
// log level that is independently controlled from parent.
//
// Note that events are filtered as the flow from their origin branch to the
// base. When a parent Logger has a more restrictive log level than a child
// Logger, the event might pass through from a child to its parent, but be
// filtered out once it arrives at the parent.
func NewBranch(parent *Logger) *Logger {
	return &Logger{parent: parent}
}

// NewBranchWithPrefix returns a new Logger instance that logs to parent, but
// has its own log level that is independently controlled from
// parent. Furthermore, events that pass through the returned Logger will have
// prefix string prefixed to the event.
//
// Note that events are filtered as the flow from their origin branch to the
// base. When a parent Logger has a more restrictive log level than a child
// Logger, the event might pass through from a child to its parent, but be
// filtered out once they arrive at the parent.
func NewBranchWithPrefix(parent *Logger, prefix string) *Logger {
	return &Logger{parent: parent, prefix: prefix}
}

// NewTracer returns a new Logger instance that sets the tracer bit for events
// that are logged to it.
//
//     tl := NewTracer(logger, "[QUERY-1234] ") // make a trace logger
//     tl.Debug("start handling: %f", 3.14)       // [QUERY-1234] start handling: 3.14
func NewTracer(parent *Logger, prefix string) *Logger {
	return &Logger{parent: parent, prefix: prefix, tracer: true}
}

func (b *Logger) log(e *event) error {
	if !e.tracer && Level(atomic.LoadUint32((*uint32)(&b.level))) > e.level {
		return nil
	}
	if b.prefix != "" {
		e.prefix = append([]string{b.prefix}, e.prefix...)
	}
	return b.parent.log(e)
}

// SetLevel allows changing the log level. Events must have the same log level
// or higher for events to be logged.
func (b *Logger) SetLevel(level Level) *Logger {
	atomic.StoreUint32((*uint32)(&b.level), uint32(level))
	return b
}

// SetDebug changes the log level to Debug, which allows all events to be
// logged.
func (b *Logger) SetDebug() *Logger {
	atomic.StoreUint32((*uint32)(&b.level), uint32(Debug))
	return b
}

// SetVerbose changes the log level to Verbose, which causes all Debug events to
// be ignored, and all Verbose, Info, Warning, and Error events to be logged.
func (b *Logger) SetVerbose() *Logger {
	atomic.StoreUint32((*uint32)(&b.level), uint32(Verbose))
	return b
}

// SetInfo changes the log level to Info, which causes all Debug and Verbose
// events to be ignored, and all Info, Warning, and Error events to be logged.
func (b *Logger) SetInfo() *Logger {
	atomic.StoreUint32((*uint32)(&b.level), uint32(Info))
	return b
}

// SetWarning changes the log level to Warning, which causes all Debug, Verbose,
// and Info events to be ignored, and all Warning, and Error events to be
// logged.
func (b *Logger) SetWarning() *Logger {
	atomic.StoreUint32((*uint32)(&b.level), uint32(Warning))
	return b
}

// SetError changes the log level to Error, which causes all Debug, Verbose,
// Info, and Warning events to be ignored, and all Error events to be logged.
func (b *Logger) SetError() *Logger {
	atomic.StoreUint32((*uint32)(&b.level), uint32(Error))
	return b
}

// Debug is used to inject a Debug event into the logs.
func (b *Logger) Debug(format string, args ...interface{}) error {
	if Level(atomic.LoadUint32((*uint32)(&b.level))) > Debug {
		return nil
	}
	var prefix []string
	if b.prefix != "" {
		prefix = []string{b.prefix}
	}
	return b.parent.log(&event{format: format, args: args, prefix: prefix, tracer: b.tracer, level: Debug})
}

// Verbose is used to inject a Verbose event into the logs.
func (b *Logger) Verbose(format string, args ...interface{}) error {
	if Level(atomic.LoadUint32((*uint32)(&b.level))) > Verbose {
		return nil
	}
	var prefix []string
	if b.prefix != "" {
		prefix = []string{b.prefix}
	}
	return b.parent.log(&event{format: format, args: args, prefix: prefix, tracer: b.tracer, level: Verbose})
}

// Info is used to inject a Info event into the logs.
func (b *Logger) Info(format string, args ...interface{}) error {
	if Level(atomic.LoadUint32((*uint32)(&b.level))) > Info {
		return nil
	}
	var prefix []string
	if b.prefix != "" {
		prefix = []string{b.prefix}
	}
	return b.parent.log(&event{format: format, args: args, prefix: prefix, tracer: b.tracer, level: Info})
}

// Warning is used to inject a Warning event into the logs.
func (b *Logger) Warning(format string, args ...interface{}) error {
	if Level(atomic.LoadUint32((*uint32)(&b.level))) > Warning {
		return nil
	}
	var prefix []string
	if b.prefix != "" {
		prefix = []string{b.prefix}
	}
	return b.parent.log(&event{format: format, args: args, prefix: prefix, tracer: b.tracer, level: Warning})
}

// Error is used to inject a Error event into the logs.
func (b *Logger) Error(format string, args ...interface{}) error {
	var prefix []string
	if b.prefix != "" {
		prefix = []string{b.prefix}
	}
	return b.parent.log(&event{format: format, args: args, prefix: prefix, tracer: b.tracer, level: Error})
}

// compileFormat converts the format string into a slice of functions to invoke
// when creating a log line.
func compileFormat(format string) ([]func(*event, *[]byte), bool, error) {
	// build slice of emitter functions, each will emit the requested
	// information
	var emitters []func(*event, *[]byte)

	// Implemented as a state machine that alternates between 2 states: either
	// capturing runes for the next constant buffer, or capturing runes for the
	// next token
	var buf, token []byte
	var indexOpenCurlyBrace int  // index of most recent open curly brace
	var isCapturingToken bool    // true after open curly brace until next close curly brace
	var isPrevRuneBackslash bool // true when previous rune was backslash
	var isPrevRuneNewline bool   // true when rune most recently read is newline
	var isTimeRequired bool      // true when any of the formatters require system time

	for ri, rune := range format {
		isPrevRuneNewline = rune == '\n'

		if isPrevRuneBackslash {
			// When this rune has been escaped, then just write it out to
			// whichever buffer we're collecting to right now.
			if isCapturingToken {
				appendRune(&token, rune)
			} else {
				appendRune(&buf, rune)
			}
			isPrevRuneBackslash = false
			continue
		}

		switch rune {
		case '\\':
			isPrevRuneBackslash = true
		case '{':
			if isCapturingToken {
				return nil, false, fmt.Errorf("cannot compile log format with embedded curly braces; runes %d and %d", indexOpenCurlyBrace, ri)
			}
			// Stop capturing buf, and begin capturing token.
			emitters = append(emitters, makeStringEmitter(string(buf)))
			buf = buf[:0]
			isCapturingToken = true
			indexOpenCurlyBrace = ri
		case '}':
			if !isCapturingToken {
				return nil, false, fmt.Errorf("cannot compile log format with unmatched closing curly braces; rune %d", ri)
			}
			// Stop capturing token, and begin capturing buf.
			switch tok := string(token); tok {
			case "epoch":
				isTimeRequired = true
				emitters = append(emitters, epochEmitter)
			case "iso8601":
				isTimeRequired = true
				emitters = append(emitters, makeUTCTimestampEmitter(time.RFC3339))
			case "level":
				emitters = append(emitters, levelEmitter)
			case "message":
				emitters = append(emitters, messageEmitter)
			case "program":
				emitters = append(emitters, makeProgramEmitter())
			case "timestamp":
				// Emulate timestamp format from stdlib log (log.LstdFlags).
				isTimeRequired = true
				emitters = append(emitters, makeUTCTimestampEmitter("2006/01/02 15:04:05"))
			default:
				// ??? Not sure how I feel about the below API.
				if strings.HasPrefix(tok, "localtime=") {
					emitters = append(emitters, makeLocalTimestampEmitter(tok[10:]))
				} else if strings.HasPrefix(tok, "utctime=") {
					emitters = append(emitters, makeUTCTimestampEmitter(tok[8:]))
				} else {
					return nil, false, fmt.Errorf("cannot compile log format with unknown formatting verb %q", token)
				}
				isTimeRequired = true
			}
			token = token[:0]
			isCapturingToken = false
		default:
			// Append rune to either token or buf.
			if isCapturingToken {
				appendRune(&token, rune)
			} else {
				appendRune(&buf, rune)
			}
		}
	}

	if isCapturingToken {
		return nil, false, fmt.Errorf("cannot compile log format with unmatched opening curly braces; rune %d", indexOpenCurlyBrace)
	}

	if !isPrevRuneNewline {
		buf = append(buf, '\n') // terminate each log line with newline byte
	}

	if len(buf) > 0 {
		emitters = append(emitters, makeStringEmitter(string(buf)))
	}

	return emitters, isTimeRequired, nil
}

func appendRune(buf *[]byte, r rune) {
	if r < utf8.RuneSelf {
		*buf = append(*buf, byte(r))
		return
	}
	olen := len(*buf)
	*buf = append(*buf, 0, 0, 0, 0)              // grow buf large enough to accommodate largest possible UTF8 sequence
	n := utf8.EncodeRune((*buf)[olen:olen+4], r) // encode rune into newly allocated buf space
	*buf = (*buf)[:olen+n]                       // trim buf to actual size used by rune addition
}

func epochEmitter(e *event, bb *[]byte) {
	*bb = append(*bb, strconv.FormatInt(e.when.UTC().Unix(), 10)...)
}

func levelEmitter(e *event, bb *[]byte) {
	*bb = append(*bb, e.level.String()...)
}

var program string

func makeProgramEmitter() func(e *event, bb *[]byte) {
	if program == "" {
		var err error
		program, err = os.Executable()
		if err != nil {
			program = os.Args[0]
		}
		program = filepath.Base(program)
	}
	return func(e *event, bb *[]byte) {
		*bb = append(*bb, program...)
	}
}

func makeStringEmitter(value string) func(*event, *[]byte) {
	return func(_ *event, bb *[]byte) {
		*bb = append(*bb, value...)
	}
}

func makeLocalTimestampEmitter(format string) func(e *event, bb *[]byte) {
	return func(e *event, bb *[]byte) {
		*bb = append(*bb, e.when.Format(format)...)
	}
}

func makeUTCTimestampEmitter(format string) func(e *event, bb *[]byte) {
	return func(e *event, bb *[]byte) {
		*bb = append(*bb, e.when.UTC().Format(format)...)
	}
}

func messageEmitter(e *event, bb *[]byte) {
	*bb = append(*bb, strings.Join(e.prefix, "")...)       // emit the event's prefix ???
	*bb = append(*bb, fmt.Sprintf(e.format, e.args...)...) // followed by the event message
}
