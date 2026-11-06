package action

import (
	"errors"
	"net/url"
	"strings"
)

// Type defines kind of action to execute when alarm fires.
type Type string

const (
	TypeNone    Type = "none"
	TypeOpenURL Type = "open_url"
)

// AlarmAction describes an automated action attached to an alarm.
type AlarmAction struct {
	Type   Type   `json:"type"`
	Target string `json:"target"` // HTTPS/HTTP URL
	Args   string `json:"args"`
}

// Validate checks safety constraints on alarm action.
func (a *AlarmAction) Validate() error {
	if a.Type == TypeNone || a.Type == "" {
		return nil
	}

	target := strings.TrimSpace(a.Target)
	if target == "" {
		return errors.New("target cannot be empty")
	}

	switch a.Type {
	case TypeOpenURL:
		parsed, err := url.Parse(target)
		if err != nil {
			return errors.New("invalid URL: " + err.Error())
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return errors.New("URL scheme must be http or https")
		}
		host := strings.ToLower(parsed.Hostname())
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return errors.New("loopback URLs not permitted for alarm actions")
		}
	default:
		return errors.New("unsupported action type")
	}

	return nil
}
