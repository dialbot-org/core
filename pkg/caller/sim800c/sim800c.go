package sim800c

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

const (
	defaultPort      = "/dev/ttyUSB0"
	defaultBaud      = 115200
	defaultAudioFile = `C:\message.wav`
)

type Outcome string

const (
	OutcomeInterested      Outcome = "interested"
	OutcomeCompleted       Outcome = "completed"
	OutcomeBusy            Outcome = "busy"
	OutcomeNoAnswer        Outcome = "no_answer"
	OutcomeNoCarrier       Outcome = "no_carrier"
	OutcomePlaybackFailed  Outcome = "playback_failed"
	OutcomeModemError      Outcome = "modem_error"
	OutcomeSerialError     Outcome = "serial_error"
)

type Config struct {
	Port      string
	BaudRate  int
	AudioFile string
}

type Client struct {
	cfg Config
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
	return &Client{cfg: cfg}
}

type CallResult struct {
	Outcome Outcome
	Detail  string
}

func (c *Client) Call(number string) CallResult {
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

	if err := write(port, "ATD"+number+";\r\n"); err != nil {
		log.Printf("sim800c: dial: %v", err)
		return CallResult{Outcome: OutcomeSerialError, Detail: err.Error()}
	}

	response := readFor(port, 3*time.Second)

	if strings.Contains(response, "BUSY") {
		return CallResult{Outcome: OutcomeBusy, Detail: "call was busy"}
	}
	if strings.Contains(response, "NO ANSWER") {
		return CallResult{Outcome: OutcomeNoAnswer, Detail: "no answer"}
	}
	if strings.Contains(response, "NO CARRIER") {
		return CallResult{Outcome: OutcomeNoCarrier, Detail: "no carrier"}
	}

	time.Sleep(5 * time.Second)

	playCommand := fmt.Sprintf(`AT+CREC=4,"%s",0,100`, c.cfg.AudioFile)
	if err := write(port, playCommand+"\r\n"); err != nil {
		log.Printf("sim800c: play command: %v", err)
		return CallResult{Outcome: OutcomeSerialError, Detail: err.Error()}
	}

	playResponse := readFor(port, 5*time.Second)
	if strings.Contains(playResponse, "ERROR") {
		return CallResult{Outcome: OutcomePlaybackFailed, Detail: playResponse}
	}

	switch listenForDTMF(port) {
	case OutcomeInterested:
		return CallResult{Outcome: OutcomeInterested, Detail: "dtmf 1"}
	case OutcomeNoCarrier:
		return CallResult{Outcome: OutcomeCompleted, Detail: "call ended without dtmf"}
	default:
		return CallResult{Outcome: OutcomeCompleted, Detail: "call ended"}
	}
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

func listenForDTMF(port serial.Port) Outcome {
	buf := make([]byte, 4096)

	for {
		n, err := port.Read(buf)
		if err == nil && n > 0 {
			response := string(buf[:n])

			if strings.Contains(response, "+DTMF: 1") {
				hangup(port)
				return OutcomeInterested
			}

			if strings.Contains(response, "NO CARRIER") {
				return OutcomeNoCarrier
			}
		}

		time.Sleep(20 * time.Millisecond)
	}
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
