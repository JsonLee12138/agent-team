package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
)

const workerProviderFlagHelp = "AI provider (claude|codex|gemini|opencode)"

func validateWorkerProvider(provider string) error {
	if provider == "" {
		return fmt.Errorf("unsupported --provider %q (supported: claude, codex, gemini, opencode)", provider)
	}
	if !supportedWorkerProvider(provider) {
		return fmt.Errorf("unsupported --provider %q (supported: claude, codex, gemini, opencode)", provider)
	}
	return nil
}

func supportedWorkerProvider(provider string) bool {
	return internal.SupportedProviders[provider]
}
