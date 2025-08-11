package railway

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// VariableDiff 表示变量差异
type VariableDiff struct {
	AddedOrUpdated map[string]string
	Removed        []string
}

// DiffVariables 计算 desired 与 current 的差异
func DiffVariables(current, desired map[string]string) VariableDiff {
	diff := VariableDiff{AddedOrUpdated: map[string]string{}, Removed: []string{}}
	// additions/updates
	for k, v := range desired {
		if cv, ok := current[k]; !ok || cv != v {
			diff.AddedOrUpdated[k] = v
		}
	}
	// removals
	for k := range current {
		if _, ok := desired[k]; !ok {
			diff.Removed = append(diff.Removed, k)
		}
	}
	sort.Strings(diff.Removed)
	return diff
}

// SerializeVariablesJSON 以 JSON 序列化变量
func SerializeVariablesJSON(vars map[string]string) ([]byte, error) {
	return json.MarshalIndent(vars, "", "  ")
}

// ParseVariablesJSON 从 JSON 解析变量
func ParseVariablesJSON(data []byte) (map[string]string, error) {
	out := map[string]string{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SerializeVariablesDotenv 以 .env 格式序列化变量
func SerializeVariablesDotenv(vars map[string]string) string {
	// 稳定顺序
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		v := vars[k]
		// 基础转义
		v = strings.ReplaceAll(v, "\n", "\\n")
		if strings.ContainsAny(v, " #\t\"") {
			v = fmt.Sprintf("\"%s\"", strings.ReplaceAll(v, "\"", "\\\""))
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
		b.WriteString("\n")
	}
	return b.String()
}

// ParseVariablesDotenv 从 .env 文本解析变量（不支持复杂插值，仅键值行）
func ParseVariablesDotenv(data []byte) (map[string]string, error) {
	out := map[string]string{}
	s := bufio.NewScanner(strings.NewReader(string(data)))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 允许 export 前缀
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		// 去掉引号
		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") && len(val) >= 2 {
			val = strings.TrimSuffix(strings.TrimPrefix(val, "\""), "\"")
			val = strings.ReplaceAll(val, "\\\"", "\"")
		}
		val = strings.ReplaceAll(val, "\\n", "\n")
		if key != "" {
			out[key] = val
		}
	}
	return out, s.Err()
}

// SaveVariablesToFile 保存变量到文件（根据扩展名选择格式：.json 或 其他为 .env）
func SaveVariablesToFile(path string, vars map[string]string) error {
	if len(vars) == 0 {
		return errors.New("no variables to save")
	}
	var b []byte
	var err error
	if strings.EqualFold(filepath.Ext(path), ".json") {
		b, err = SerializeVariablesJSON(vars)
	} else {
		s := SerializeVariablesDotenv(vars)
		b = []byte(s)
	}
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

// LoadVariablesFromFile 从文件加载变量（支持 .json 与 .env）
func LoadVariablesFromFile(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(filepath.Ext(path), ".json") {
		return ParseVariablesJSON(b)
	}
	return ParseVariablesDotenv(b)
}

// ApplyVariableDiff 根据 replace 策略应用 desired 相对 current 的差异
// - replace=false: 仅上送新增/更新键
// - replace=true: 覆盖式上送 desired 集合（将移除 current 中不存在的键）
func (c *Client) ApplyVariableDiff(ctx context.Context, projectID, environmentID string, serviceID *string, replace bool, current, desired map[string]string) error {
	if replace {
		return c.UpsertVariables(ctx, projectID, environmentID, serviceID, true, desired)
	}
	diff := DiffVariables(current, desired)
	if len(diff.AddedOrUpdated) == 0 {
		return nil
	}
	return c.UpsertVariables(ctx, projectID, environmentID, serviceID, false, diff.AddedOrUpdated)
}
