package hostsmgr

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const hostsPath = "C:\\Windows\\System32\\drivers\\etc\\hosts"

type Manager struct {
	Path string
}

func New() *Manager {
	return &Manager{Path: hostsPath}
}

// Add 添加域名重定向：127.0.0.1 <domain>
func (m *Manager) Add(domain string) error {
	entry := fmt.Sprintf("127.0.0.1 %s", domain)
	content, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Errorf("read hosts: %w", err)
	}
	if strings.Contains(string(content), entry) {
		return nil // 已存在
	}
	f, err := os.OpenFile(m.Path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open hosts: %w", err)
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "\n# ComeHere\n%s\n", entry); err != nil {
		return fmt.Errorf("write hosts: %w", err)
	}
	return nil
}

// Remove 移除域名重定向
func (m *Manager) Remove(domain string) error {
	content, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Errorf("read hosts: %w", err)
	}
	lines := strings.Split(string(content), "\n")
	var newLines []string
	inComeHereBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "# ComeHere" {
			inComeHereBlock = true
			continue
		}
		if inComeHereBlock {
			inComeHereBlock = false
			if strings.Contains(trimmed, domain) {
				continue // 跳过这条
			}
			// 如果不是当前域名，保留
			newLines = append(newLines, "# ComeHere")
		}
		if strings.Contains(trimmed, "127.0.0.1 "+domain) {
			continue
		}
		newLines = append(newLines, line)
	}
	return os.WriteFile(m.Path, []byte(strings.Join(newLines, "\n")), 0644)
}

// HasDomain 检查域名是否已被劫持
func (m *Manager) HasDomain(domain string) (bool, error) {
	f, err := os.Open(m.Path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "127.0.0.1 "+domain) {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// ListManaged 列出所有被 ComeHere 管理的域名
func (m *Manager) ListManaged() ([]string, error) {
	f, err := os.Open(m.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var domains []string
	inComeHereBlock := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "# ComeHere" {
			inComeHereBlock = true
			continue
		}
		if inComeHereBlock && strings.Contains(line, "127.0.0.1 ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				domains = append(domains, parts[1])
			}
		}
		inComeHereBlock = false
	}
	return domains, scanner.Err()
}

// Cleanup 清除所有 ComeHere 添加的 hosts 条目
func (m *Manager) Cleanup() error {
	content, err := os.ReadFile(m.Path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	var newLines []string
	skip := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "# ComeHere" {
			skip = true
			continue
		}
		if skip {
			skip = false
			continue
		}
		newLines = append(newLines, line)
	}
	return os.WriteFile(m.Path, []byte(strings.Join(newLines, "\n")), 0644)
}
