package config

import (
	"encoding/json"
	"sync"

	"golang.org/x/text/language"
)

// Bundle 管理所有语言的翻译资源，并发安全。
type Bundle struct {
	mu       sync.RWMutex
	locales  map[string]map[string]string // lang → msgID → text
	defaults map[string]string            // 英文作为 fallback
}

func NewBundle() *Bundle {
	return &Bundle{
		locales:  make(map[string]map[string]string),
		defaults: make(map[string]string),
	}
}

// Load 加载指定语言的翻译 JSON。data 是 {"MSG_ID": "翻译文本"} 格式。
func (b *Bundle) Load(lang string, data []byte) error {
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.locales[lang] = m

	// 默认语言用英文
	if lang == "en" {
		b.defaults = m
	}

	return nil
}

// Translate 根据 Accept-Language 头返回翻译后的文本。
// 匹配不到时 fallback 到英文，英文也没有就返回 raw 原文。
func (b *Bundle) Translate(acceptLanguage, key, raw string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	lang := parseAcceptLanguage(acceptLanguage)

	if m, ok := b.locales[lang]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}

	if s, ok := b.defaults[key]; ok {
		return s
	}

	return raw
}

// parseAcceptLanguage 从 Accept-Language 头提取首选语言标签。
// 只取第一个，类似 "zh-CN,en;q=0.9" → "zh"。
func parseAcceptLanguage(header string) string {
	if header == "" {
		return "en"
	}
	tag, _, err := language.ParseAcceptLanguage(header)
	if err != nil || len(tag) == 0 {
		return "en"
	}
	base, _ := tag[0].Base()
	return base.String()
}

// ── 全局单例 —— 简单场景够用 ──────────────────────────────

var globalBundle = NewBundle()

func InitBundle() *Bundle {
	return globalBundle
}

func LoadLocale(lang string, data []byte) error {
	return globalBundle.Load(lang, data)
}

// Translate 全局快捷函数，供 response 包调用。
func Translate(acceptLanguage, key, raw string) string {
	return globalBundle.Translate(acceptLanguage, key, raw)
}

// DetectLang 是 Translate 的别名，便于 validator 等包使用。
func DetectLang(header string) string {
	return parseAcceptLanguage(header)
}
