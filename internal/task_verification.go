package internal

import (
	"fmt"
	"os"
	"strings"
)

// VerificationResult represents the parsed verification gate result.
type VerificationResult string

const (
	VerificationResultMissing VerificationResult = "missing"
	VerificationResultPending VerificationResult = "pending"
	VerificationResultPass    VerificationResult = "pass"
	VerificationResultPartial VerificationResult = "partial"
	VerificationResultFail    VerificationResult = "fail"
)

func ReadTaskVerificationResult(root, taskID string, location TaskRecordLocation) (VerificationResult, error) {
	path := taskVerificationPathByLocation(root, taskID, location)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return VerificationResultMissing, nil
		}
		return "", fmt.Errorf("read verification.md: %w", err)
	}
	return ParseVerificationResult(string(data)), nil
}

func ParseVerificationResult(content string) VerificationResult {
	lines := strings.Split(content, "\n")
	inResult := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "## ") {
			if strings.EqualFold(line, "## Result") {
				inResult = true
				continue
			}
			if inResult {
				break
			}
		}
		if !inResult {
			continue
		}
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		value := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "- ")))
		switch VerificationResult(value) {
		case VerificationResultPending, VerificationResultPass, VerificationResultPartial, VerificationResultFail:
			return VerificationResult(value)
		default:
			return VerificationResultPending
		}
	}
	return VerificationResultPending
}

func ValidateArchiveReadiness(result VerificationResult, strict bool) error {
	switch result {
	case VerificationResultPass:
		return nil
	case VerificationResultPartial:
		if strict {
			return fmt.Errorf("strict mode only allows verification result 'pass'")
		}
		return nil
	case VerificationResultMissing:
		return fmt.Errorf("missing verification artifact")
	case VerificationResultFail:
		return fmt.Errorf("verification failed; archive is blocked")
	case VerificationResultPending:
		fallthrough
	default:
		return fmt.Errorf("verification is still pending")
	}
}

func taskVerificationPathByLocation(root, taskID string, location TaskRecordLocation) string {
	switch location {
	case TaskRecordLocationArchived:
		return TaskArchiveVerificationPath(root, taskID)
	case TaskRecordLocationDeprecated:
		return TaskDeprecatedVerificationPath(root, taskID)
	default:
		return TaskVerificationPath(root, taskID)
	}
}

func ArchiveReadyLabel(result VerificationResult) string {
	switch result {
	case VerificationResultPass:
		return "yes"
	case VerificationResultPartial:
		return "strict-no"
	default:
		return "no"
	}
}
