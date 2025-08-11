package util

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

// PromptText 提示用户输入文本
func PromptText(message string) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return "", err
	}

	return result, nil
}

// PromptPassword 提示用户输入密码
func PromptPassword(message string) (string, error) {
	var result string
	prompt := &survey.Password{
		Message: message,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return "", err
	}

	return result, nil
}

// PromptConfirm 提示用户确认
func PromptConfirm(message string) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return false, err
	}

	return result, nil
}

// PromptSelect 提示用户从列表中选择
func PromptSelect(message string, options []string) (string, error) {
	var result string
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return "", err
	}

	return result, nil
}

// PromptMultiSelect 提示用户从列表中多选
func PromptMultiSelect(message string, options []string) ([]string, error) {
	var result []string
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// PrintSuccess 打印成功消息
func PrintSuccess(message string) {
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	fmt.Printf("✅ %s\n", green(message))
}

// PrintError 打印错误消息
func PrintError(message string) {
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	fmt.Printf("❌ %s\n", red(message))
}

// PrintWarning 打印警告消息
func PrintWarning(message string) {
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	fmt.Printf("⚠️  %s\n", yellow(message))
}

// PrintInfo 打印信息消息
func PrintInfo(message string) {
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	fmt.Printf("ℹ️  %s\n", cyan(message))
}
