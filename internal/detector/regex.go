package detector

import (
	"path/filepath"
	"regexp"
	"strings"
)

type CompileRule struct {
	Rule   Rule
	Regexp *regexp.Regexp
}

type LineMatch struct {
	Rule  Rule
	Match string
}

type RegexEngine struct {
	compiled []CompileRule
}

func NewRegexEngine(rules []Rule) (*RegexEngine, error) {
	engine := &RegexEngine{}
	for _, rule := range rules {
		if rule.Pattern == "" {
			continue
		}
		// 编译正则表达式
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			// fmt.Printf("[调试] 规则编译失败 %s: %v\n", rule.Name, err)
			continue
		}
		engine.compiled = append(engine.compiled, CompileRule{
			Rule:   rule,
			Regexp: re,
		})
	}
	// fmt.Printf("[调试] 已加载 %d 条规则\n", len(engine.compiled))
	return engine, nil
}

func (e *RegexEngine) MatchLine(line, filePath string) []LineMatch {
	var matches []LineMatch
	ext := strings.ToLower(filepath.Ext(filePath))

	for _, cr := range e.compiled {
		// 有扩展名限制才检查
		if len(cr.Rule.Extensions) > 0 && !containsExt(cr.Rule.Extensions, ext) {
			continue
		}

		// 查找所有匹配
		if cr.Regexp.MatchString(line) {
			// 获取实际匹配的字符串
			matchStr := cr.Regexp.FindString(line)
			matches = append(matches, LineMatch{
				Rule:  cr.Rule,
				Match: matchStr,
			})
		}
	}
	return matches
}

func containsExt(exts []string, ext string) bool {
	for _, e := range exts {
		if strings.EqualFold(e, ext) {
			return true
		}
	}
	return false
}
