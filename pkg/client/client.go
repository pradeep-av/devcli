package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
)

// RequestOptions contains settings for making a REST request.
type RequestOptions struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
	Verbose bool
}

// Send executes the HTTP request, measures latency, and prints formatted output to the terminal.
func Send(opt RequestOptions) error {
	var bodyReader io.Reader
	if opt.Body != "" {
		bodyReader = strings.NewReader(opt.Body)
	}

	req, err := http.NewRequest(opt.Method, opt.URL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Apply headers
	for k, v := range opt.Headers {
		req.Header.Set(k, v)
	}

	// Default Content-Type if body exists and header is empty
	if opt.Body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Verbose request details
	if opt.Verbose {
		color.New(color.Bold, color.FgCyan).Println(">>> REQUEST >>>")
		color.New(color.FgCyan).Printf("%s %s\n", opt.Method, opt.URL)
		for k, vv := range req.Header {
			color.New(color.FgCyan).Printf("  %s: %s\n", k, strings.Join(vv, ", "))
		}
		if opt.Body != "" {
			color.New(color.FgCyan).Println("\n" + opt.Body)
		}
		fmt.Println()
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Format response header
	color.New(color.Bold, color.FgCyan).Println("<<< RESPONSE <<<")

	var statusColor *color.Color
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		statusColor = color.New(color.FgGreen, color.Bold)
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		statusColor = color.New(color.FgYellow, color.Bold)
	default:
		statusColor = color.New(color.FgRed, color.Bold)
	}

	fmt.Printf("Status: ")
	statusColor.Printf("%d %s", resp.StatusCode, resp.Status)
	fmt.Printf(" | Duration: ")
	color.New(color.FgMagenta).Printf("%v", duration)
	fmt.Printf(" | Size: ")
	color.New(color.FgBlue).Printf("%d bytes\n", len(respBody))

	if opt.Verbose {
		for k, vv := range resp.Header {
			color.New(color.FgHiBlack).Printf("  %s: %s\n", k, strings.Join(vv, ", "))
		}
		fmt.Println()
	}

	if len(respBody) > 0 {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
			fmt.Println(ColorizeJSON(prettyJSON.String()))
		} else {
			fmt.Println(string(respBody))
		}
	} else {
		color.New(color.Italic, color.FgHiBlack).Println("(empty response body)")
	}

	return nil
}

// ColorizeJSON adds terminal colors to JSON keys, strings, booleans, and numbers.
func ColorizeJSON(jsonStr string) string {
	var result strings.Builder
	inQuote := false
	escaped := false

	keyColor := color.New(color.FgHiBlue, color.Bold).SprintFunc()
	valStrColor := color.New(color.FgGreen).SprintFunc()
	valNumColor := color.New(color.FgYellow).SprintFunc()
	valBoolColor := color.New(color.FgCyan).SprintFunc()
	syntaxColor := color.New(color.FgHiBlack).SprintFunc()

	var token strings.Builder

	flushToken := func() {
		if token.Len() == 0 {
			return
		}
		s := token.String()
		token.Reset()

		if s == "true" || s == "false" || s == "null" {
			result.WriteString(valBoolColor(s))
		} else if isNumber(s) {
			result.WriteString(valNumColor(s))
		} else {
			result.WriteString(s)
		}
	}

	for i := 0; i < len(jsonStr); i++ {
		ch := jsonStr[i]

		if escaped {
			token.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			token.WriteByte(ch)
			escaped = true
			continue
		}

		if ch == '"' {
			if inQuote {
				inQuote = false
				s := token.String()
				token.Reset()

				// Lookahead to check if this token is a JSON key (followed by colon ':')
				nextIsColon := false
				for j := i + 1; j < len(jsonStr); j++ {
					nextCh := jsonStr[j]
					if nextCh == ' ' || nextCh == '\t' || nextCh == '\n' || nextCh == '\r' {
						continue
					}
					if nextCh == ':' {
						nextIsColon = true
					}
					break
				}

				if nextIsColon {
					result.WriteString(keyColor(`"` + s + `"`))
				} else {
					result.WriteString(valStrColor(`"` + s + `"`))
				}
			} else {
				flushToken()
				inQuote = true
			}
			continue
		}

		if inQuote {
			token.WriteByte(ch)
			continue
		}

		// Non-quote context
		if ch == '{' || ch == '}' || ch == '[' || ch == ']' || ch == ',' || ch == ':' {
			flushToken()
			result.WriteString(syntaxColor(string(ch)))
		} else if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			flushToken()
			result.WriteByte(ch)
		} else {
			token.WriteByte(ch)
		}
	}
	flushToken()

	return result.String()
}

func isNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && c != '.' && c != '-' && c != '+' && c != 'e' && c != 'E' {
			return false
		}
	}
	return true
}
