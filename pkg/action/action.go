// Package action 定义闹钟触发时可执行的安全自动化动作（如打开外部 HTTP/HTTPS 链接）及其安全校验规则。
package action

import (
	"errors"
	"net/url"
	"strings"
)

// Type 定义闹钟触发时执行的动作类型。
// Type defines kind of action to execute when alarm fires.
type Type string

const (
	TypeNone    Type = "none"     // 无附加动作
	TypeOpenURL Type = "open_url" // 打开指定的网页链接
)

// AlarmAction 描述绑定在闹钟上的自动化触发动作。
// AlarmAction describes an automated action attached to an alarm.
type AlarmAction struct {
	Type   Type   `json:"type"`   // 动作类型（none 或 open_url）
	Target string `json:"target"` // 目标地址（仅允许安全的 HTTP/HTTPS URL）
	Args   string `json:"args"`   // 附加参数
}

// Validate 校验闹钟附加动作的安全性约束（仅允许 http/https 协议，并拦截本地环回地址以防 SSRF/本地服务探测）。
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
		// 仅允许标准网页协议
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return errors.New("URL scheme must be http or https")
		}
		// 禁止访问本地回环地址
		host := strings.ToLower(parsed.Hostname())
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return errors.New("loopback URLs not permitted for alarm actions")
		}
	default:
		return errors.New("unsupported action type")
	}

	return nil
}
