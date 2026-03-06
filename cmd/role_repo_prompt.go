package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func isInteractiveInput(in io.Reader) bool {
	f, ok := in.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func promptSelectNames(in io.Reader, out io.Writer, title string, options []string) ([]string, error) {
	if len(options) == 0 {
		return []string{}, nil
	}
	if !isInteractiveInput(in) {
		return options, nil
	}
	selected := []string{}
	prompt := buildRoleRepoMultiSelect(title, options)
	if err := survey.AskOne(prompt, &selected, survey.WithValidator(survey.Required)); err != nil {
		if err == terminal.InterruptErr {
			return nil, fmt.Errorf("selection cancelled")
		}
		return nil, err
	}
	return selected, nil
}

func buildRoleRepoMultiSelect(title string, options []string) *survey.MultiSelect {
	return &survey.MultiSelect{
		Message: title,
		Options: options,
		Default: []string{},
	}
}

func promptSingleChoice(in io.Reader, out io.Writer, message string, options []string, defaultChoice string) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options provided")
	}
	if !isInteractiveInput(in) {
		return defaultChoice, nil
	}
	selected := defaultChoice
	prompt := &survey.Select{
		Message: message,
		Options: options,
		Default: defaultChoice,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		if err == terminal.InterruptErr {
			return "", fmt.Errorf("selection cancelled")
		}
		return "", err
	}
	return selected, nil
}

func promptConfirm(in io.Reader, out io.Writer, message string) (bool, error) {
	selected, err := promptSingleChoice(in, out, message, []string{"Yes", "No"}, "No")
	if err != nil {
		return false, err
	}
	return selected == "Yes", nil
}
