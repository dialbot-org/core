package sim800c

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

const (
	defaultPort      = "/dev/ttyUSB0"
	defaultBaud      = 115200
	defaultAudioFile = `C:\message.amr`

	dialResponseWait = 3 * time.Second
	maxRingTime      = 60 * time.Second
	maxListenTime    = 5 * time.Minute
)

type Outcome string

const (
	OutcomeInterested     Outcome = "interested"
	OutcomeCompleted      Outcome = "completed"
	OutcomeBusy           Outcome = "busy"
	OutcomeNoAnswer       Outcome = "no_answer"
	OutcomeNoCarrier      Outcome = "no_carrier"
	OutcomePlaybackFailed Outcome = "playback_failed"
	OutcomeModemError     Outcome = "modem_error"
	OutcomeSerialError    Outcome = "serial_error"
	OutcomeCallTimeout    Outcome = "call_timeout"
)

type Config struct {
	Port          string
	BaudRate      int
	AudioFile     string
	PlaybackDelay int
}

type Client struct {
	cfg       Config
	callMutex sync.Mutex
}

func New(cfg Config) *Client {
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	if cfg.BaudRate == 0 {
		cfg.BaudRate = defaultBaud
	}
	if cfg.AudioFile == "" {
		cfg.AudioFile = defaultAudioFile
	}
	if cfg.PlaybackDelay <= 0 {
		cfg.PlaybackDelay = 2
	}
	return &Client{cfg: cfg}
}

type CallResult struct {
	Outcome Outcome
	Detail  string
}

func (c *Client) Call(number string, callerName string) CallResult {
	c.callMutex.Lock()
	defer c.callMutex.Unlock()

	port, err := serial.Open(c.cfg.Port, &serial.Mode{
		BaudRate: c.cfg.BaudRate,
	})
	if err != nil {
		log.Printf("sim800c: open serial port: %v", err)
		return CallResult{Outcome: OutcomeSerialError, Detail: err.Error()}
	}
	defer port.Close()

	port.SetReadTimeout(100 * time.Millisecond)

	if err := command(port, "AT+DDET=1,0,0"); err != nil {
		log.Printf("sim800c: AT+DDET: %v", err)
		return CallResult{Outcome: OutcomeModemError, Detail: err.Error()}
	}

	if err := command(port, "AT+DTAM=1"); err != nil {
		log.Printf("sim800c: AT+DTAM: %v", err)
		return CallResult{Outcome: OutcomeModemError, Detail: err.Error()}
	}

	dial := formatDialNumber(number)
	if dial == "" {
		return CallResult{Outcome: OutcomeModemError, Detail: "empty or invalid phone number"}
	}

	log.Printf("sim800c: dialing %q for %q (stored as %q)", dial, callerName, number)

	if err := write(port, "ATD"+dial+";\r\n"); err != nil {
		log.Printf("sim800c: dial: %v", err)
		return CallResult{Outcome: OutcomeSerialError, Detail: err.Error()}
	}

	response := readFor(port, dialResponseWait)
	log.Printf("sim800c: dial response: %q", strings.TrimSpace(response))

	if result, ok := parseDialResponse(response); ok {
		hangup(port)
		return result
	}

	if result, ok := waitForActiveCall(port, maxRingTime); ok {
		hangup(port)
		return result
	}

	playbackWait := time.Duration(c.cfg.PlaybackDelay) * time.Second
	if playbackWait > 0 {
		log.Printf("sim800c: call answered; waiting %s before playback", playbackWait)
		time.Sleep(playbackWait)
	}

	playCommand := fmt.Sprintf(`AT+CREC=4,"%s",0,100`, c.cfg.AudioFile)
	if err := write(port, playCommand+"\r\n"); err != nil {
		log.Printf("sim800c: play command: %v", err)
		return CallResult{Outcome: OutcomeSerialError, Detail: err.Error()}
	}

	playResponse := readFor(port, 5*time.Second)
	if strings.Contains(playResponse, "ERROR") {
		hangup(port)
		return CallResult{Outcome: OutcomePlaybackFailed, Detail: playResponse}
	}
	if containsDTMFDigit(playResponse) {
		hangup(port)
		return CallResult{Outcome: OutcomeInterested, Detail: "dtmf digit"}
	}
	if strings.Contains(playResponse, "NO CARRIER") {
		return CallResult{Outcome: OutcomeCompleted, Detail: "call ended without dtmf"}
	}

	deadline := time.Now().Add(maxListenTime)
	switch listenForDTMF(port, deadline) {
	case OutcomeInterested:
		return CallResult{Outcome: OutcomeInterested, Detail: "dtmf digit"}
	case OutcomeNoCarrier:
		return CallResult{Outcome: OutcomeCompleted, Detail: "call ended without dtmf"}
	case OutcomeCallTimeout:
		hangup(port)
		return CallResult{Outcome: OutcomeCallTimeout, Detail: "timed out waiting for call end or dtmf"}
	default:
		hangup(port)
		return CallResult{Outcome: OutcomeCompleted, Detail: "call ended"}
	}
}

// formatDialNumber keeps an Indian national number for the SIM800C.
// Country codes and trunk prefixes are dropped by taking the last 10 digits.
func formatDialNumber(number string) string {
	digits := digitsOnly(number)
	if len(digits) > 10 {
		digits = digits[len(digits)-10:]
	}
	return digits
}

func digitsOnly(number string) string {
	var b strings.Builder
	for _, ch := range strings.TrimSpace(number) {
		if ch >= '0' && ch <= '9' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func parseDialResponse(response string) (CallResult, bool) {
	upper := strings.ToUpper(response)
	if strings.Contains(upper, "ERROR") {
		return CallResult{Outcome: OutcomeModemError, Detail: strings.TrimSpace(response)}, true
	}
	if strings.Contains(upper, "NO DIALTONE") {
		return CallResult{Outcome: OutcomeNoCarrier, Detail: "no dialtone"}, true
	}
	if strings.Contains(upper, "BUSY") {
		return CallResult{Outcome: OutcomeBusy, Detail: "call was busy"}, true
	}
	if strings.Contains(upper, "NO ANSWER") {
		return CallResult{Outcome: OutcomeNoAnswer, Detail: "no answer"}, true
	}
	if strings.Contains(upper, "NO CARRIER") {
		return CallResult{Outcome: OutcomeNoCarrier, Detail: "no carrier"}, true
	}
	return CallResult{}, false
}

func waitForActiveCall(port serial.Port, timeout time.Duration) (CallResult, bool) {
	deadline := time.Now().Add(timeout)
	callSeen := false

	for time.Now().Before(deadline) {
		response, err := sendAndRead(port, "AT+CLCC", 500*time.Millisecond)
		if err != nil {
			return CallResult{Outcome: OutcomeSerialError, Detail: err.Error()}, true
		}
		if result, ok := parseDialResponse(response); ok {
			return result, true
		}

		active, present := parseCallState(response)
		if active {
			return CallResult{}, false
		}
		if present {
			callSeen = true
		} else if callSeen {
			return CallResult{Outcome: OutcomeNoAnswer, Detail: "no answer"}, true
		}

		time.Sleep(250 * time.Millisecond)
	}

	return CallResult{Outcome: OutcomeNoAnswer, Detail: "no answer"}, true
}

func parseCallState(response string) (active bool, present bool) {
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "+CLCC:") {
			continue
		}

		fields := strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "+CLCC:")), ",")
		if len(fields) < 3 {
			continue
		}

		// AT+CLCC status 0 means the call is active (the recipient answered).
		return strings.TrimSpace(fields[2]) == "0", true
	}
	return false, false
}

func command(port serial.Port, cmd string) error {
	response, err := sendAndRead(port, cmd, 2*time.Second)
	if err != nil {
		return err
	}
	if strings.Contains(response, "ERROR") {
		return fmt.Errorf("modem error: %q", response)
	}
	return nil
}

func sendAndRead(port serial.Port, cmd string, timeout time.Duration) (string, error) {
	if err := write(port, cmd+"\r\n"); err != nil {
		return "", err
	}
	return readFor(port, timeout), nil
}

func write(port serial.Port, data string) error {
	_, err := port.Write([]byte(data))
	return err
}

func containsDTMFDigit(response string) bool {
	for digit := '0'; digit <= '9'; digit++ {
		if strings.Contains(response, "+DTMF: "+string(digit)) {
			return true
		}
	}
	return false
}

func listenForDTMF(port serial.Port, deadline time.Time) Outcome {
	buf := make([]byte, 4096)

	for time.Now().Before(deadline) {
		n, err := port.Read(buf)
		if err == nil && n > 0 {
			response := string(buf[:n])

			if containsDTMFDigit(response) {
				hangup(port)
				return OutcomeInterested
			}

			if strings.Contains(response, "NO CARRIER") {
				return OutcomeNoCarrier
			}
		}

		time.Sleep(20 * time.Millisecond)
	}

	return OutcomeCallTimeout
}

func hangup(port serial.Port) {
	if err := write(port, "ATH\r\n"); err != nil {
		log.Printf("sim800c: hangup: %v", err)
	}
}

func readFor(port serial.Port, duration time.Duration) string {
	var buffer bytes.Buffer
	buf := make([]byte, 4096)

	deadline := time.Now().Add(duration)

	for time.Now().Before(deadline) {
		n, err := port.Read(buf)
		if err == nil && n > 0 {
			buffer.Write(buf[:n])
		}
		time.Sleep(10 * time.Millisecond)
	}

	return buffer.String()
}
