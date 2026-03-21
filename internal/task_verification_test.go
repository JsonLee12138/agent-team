package internal

import "testing"

func TestParseVerificationResult(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    VerificationResult
	}{
		{name: "pass", content: "# Verification\n\n## Result\n- pass\n", want: VerificationResultPass},
		{name: "partial", content: "# Verification\n\n## Result\n- partial\n", want: VerificationResultPartial},
		{name: "pending", content: "# Verification\n\n## Result\n- pending\n", want: VerificationResultPending},
		{name: "fail", content: "# Verification\n\n## Result\n- fail\n", want: VerificationResultFail},
		{name: "invalid falls back", content: "# Verification\n\n## Result\n- weird\n", want: VerificationResultPending},
		{name: "missing result falls back", content: "# Verification\n\n## Checks\n- none\n", want: VerificationResultPending},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseVerificationResult(tt.content); got != tt.want {
				t.Fatalf("ParseVerificationResult() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestValidateArchiveReadiness(t *testing.T) {
	tests := []struct {
		name    string
		result  VerificationResult
		strict  bool
		wantErr bool
	}{
		{name: "pass default", result: VerificationResultPass},
		{name: "pass strict", result: VerificationResultPass, strict: true},
		{name: "partial default", result: VerificationResultPartial},
		{name: "partial strict blocked", result: VerificationResultPartial, strict: true, wantErr: true},
		{name: "pending blocked", result: VerificationResultPending, wantErr: true},
		{name: "fail blocked", result: VerificationResultFail, wantErr: true},
		{name: "missing blocked", result: VerificationResultMissing, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArchiveReadiness(tt.result, tt.strict)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateArchiveReadiness() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
