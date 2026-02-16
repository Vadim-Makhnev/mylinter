package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v3"
)

var targetMethods = map[string]bool{
	"Debug": true,
	"Info":  true,
	"Warn":  true,
	"Error": true,
}

var targetPackages = map[string]bool{
	"log/slog":        true,
	"go.uber.org/zap": true,
}

var defaultSensitiveKeywords = []string{
	"password",
	"passwd",
	"pwd",
	"apikey",
	"token",
	"secret",
	"privatekey",
	"clientsecret",
	"authorization",
	"auth",
}

type fileConfig struct {
	SensitiveKeywords sensitiveKeywordsConfig `yaml:"sensitive_keywords"`
}

type sensitiveKeywordsConfig struct {
	Values []string `yaml:"values"`
	Add    []string `yaml:"add"`
	Remove []string `yaml:"remove"`
}

func run(pass *analysis.Pass, settings Settings) (any, error) {
	sensitiveKeywords, err := resolveSensitiveKeywords(settings.ConfigPath)
	if err != nil {
		return nil, err
	}

	// Обходим все файлы пакета и ищем вызовы функций.
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {

			switch n := n.(type) {
			// Проверяем узел на причастность к CallExpr
			case *ast.CallExpr:
				// Отбрасываем все вызовы, которые не относятся к целевым логгерам.
				if ok := isTargetLogCall(pass, n); !ok {
					return true
				}

				// Проверяем первый аргумент на чувствительные данные.
				if len(n.Args) > 0 && containsSensitiveData(n.Args[0], sensitiveKeywords) {
					pass.Reportf(n.Pos(), "log message should not contain sensitive data")
				}

				// Берем первый строковый литерал сообщения.
				str, ok := getFirstStringLiteral(pass, n)
				if !ok {
					return true
				}

				// Убираем пробелы по краям перед валидацией текста.
				msg := strings.TrimSpace(str)

				// Разрешаем только английские буквы и пробелы.
				if !isEnglishOnly(msg) {
					pass.Reportf(n.Pos(), "string should be contains only english chars")
					return true
				}

				// Первая буква сообщения должна быть в нижнем регистре.
				if !isFirstLower(msg) {
					pass.Reportf(n.Pos(), "first element should be a lower case char: %s", string([]rune(strings.TrimSpace(str))[0]))
					return true
				}
			default:
				return true
			}

			return true
		})
	}
	return nil, nil
}

func isTargetLogCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	// Берем селектор вызова: logger.Info(...)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Проверяем, что имя метода входит в целевой список.
	methodName := sel.Sel.Name
	if !targetMethods[methodName] {
		return false
	}

	// Ветка для методов на типе (например, *zap.Logger).
	selection := pass.TypesInfo.Selections[sel]
	if selection != nil {
		// Нас интересует только прямой вызов метода у значения.
		if selection.Kind() != types.MethodVal {
			return false
		}

		// Проверяем пакет метода, чтобы не ловить одноименные локальные методы.
		method, ok := selection.Obj().(*types.Func)
		if !ok || method.Pkg() == nil {
			return false
		}

		return isTargetPackage(method.Pkg().Path())
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	// Ветка для пакетных вызовов (например, slog.Info(...)).
	pkgName, ok := pass.TypesInfo.Uses[pkgIdent].(*types.PkgName)
	if !ok {
		return false
	}

	return isTargetPackage(pkgName.Imported().Path())
}

func getFirstStringLiteral(pass *analysis.Pass, call *ast.CallExpr) (string, bool) {
	// Лог-вызов без аргументов считаем ошибкой.
	if len(call.Args) == 0 {
		pass.Reportf(call.Pos(), "log call with no arguments or nil value")
		return "", false
	}

	// Берем первый аргумент функции
	firstArg := call.Args[0]

	// Пытаемся получить types.Type
	argType := pass.TypesInfo.TypeOf(firstArg)
	if argType == nil {
		return "", false
	}

	// Теперь сравниваем два type.Types
	// Проверка на тип string
	if !types.Identical(argType, types.Typ[types.String]) {
		return "", false
	}

	// Принимаем только строковый литерал, а не переменную/выражение.
	lit, ok := firstArg.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	str, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}

	if strings.TrimSpace(str) == "" {
		pass.Reportf(call.Pos(), "message should be not empty")
		return "", false
	}

	return str, true
}

func isEnglishOnly(str string) bool {
	// Пустое сообщение считаем невалидным.
	if str == "" {
		return false
	}

	// Разрешаем только латиницу и пробелы.
	for _, sym := range str {
		if !(sym >= 'a' && sym <= 'z' || sym >= 'A' && sym <= 'Z' || sym == ' ') {
			return false
		}
	}

	return true
}

func isFirstLower(str string) bool {
	// Пустую строку не проверяем на регистр.
	if str == "" {
		return false
	}

	// Проверяем первый символ после trim в вызывающем коде.
	return unicode.IsLower([]rune(str)[0])
}

func containsSensitiveData(expr ast.Expr, sensitiveKeywords []string) bool {
	// Флаг раннего выхода, если уже нашли чувствительное слово.
	found := false

	// Обходим все узлы выражения первого аргумента.
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}

		switch n := n.(type) {
		case *ast.Ident:
			// Проверяем названия идентификаторов: password, token и т.д.
			if containsSensitiveKeywordWithList(n.Name, sensitiveKeywords) {
				found = true
				return false
			}
		case *ast.BasicLit:
			// Нас интересуют только строковые литералы.
			if n.Kind != token.STRING {
				return true
			}

			str, err := strconv.Unquote(n.Value)
			if err != nil {
				return true
			}

			if containsSensitiveKeywordWithList(str, sensitiveKeywords) {
				found = true
				return false
			}
		}

		return true
	})

	return found
}

func containsSensitiveKeyword(raw string) bool {
	return containsSensitiveKeywordWithList(raw, defaultSensitiveKeywords)
}

func containsSensitiveKeywordWithList(raw string, sensitiveKeywords []string) bool {
	// Нормализуем вход перед поиском ключевых слов.
	normalized := normalizeSensitiveInput(raw)

	for _, keyword := range sensitiveKeywords {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}

	return false
}

func normalizeSensitiveInput(raw string) string {
	// Приводим к нижнему регистру и убираем типовые разделители.
	lower := strings.ToLower(raw)
	replacer := strings.NewReplacer("_", "", "-", "", " ", "", ".", "", ":", "")
	return replacer.Replace(lower)
}

func resolveSensitiveKeywords(configPath string) ([]string, error) {
	defaults := normalizeAndDeduplicateKeywords(defaultSensitiveKeywords)
	if strings.TrimSpace(configPath) == "" {
		return defaults, nil
	}

	cfg, err := readConfig(configPath)
	if err != nil {
		return nil, err
	}

	base := defaults
	if len(cfg.SensitiveKeywords.Values) > 0 {
		base = normalizeAndDeduplicateKeywords(cfg.SensitiveKeywords.Values)
	}

	removeSet := make(map[string]struct{})
	for _, keyword := range normalizeAndDeduplicateKeywords(cfg.SensitiveKeywords.Remove) {
		removeSet[keyword] = struct{}{}
	}

	resolved := make([]string, 0, len(base))
	exists := make(map[string]struct{})

	for _, keyword := range base {
		if _, skip := removeSet[keyword]; skip {
			continue
		}
		if _, ok := exists[keyword]; ok {
			continue
		}
		exists[keyword] = struct{}{}
		resolved = append(resolved, keyword)
	}

	for _, keyword := range normalizeAndDeduplicateKeywords(cfg.SensitiveKeywords.Add) {
		if _, ok := exists[keyword]; ok {
			continue
		}
		exists[keyword] = struct{}{}
		resolved = append(resolved, keyword)
	}

	return resolved, nil
}

func readConfig(configPath string) (fileConfig, error) {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return fileConfig{}, fmt.Errorf("failed to read config file %q: %w", configPath, err)
	}

	var cfg fileConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return fileConfig{}, fmt.Errorf("failed to parse config file %q: %w", configPath, err)
	}

	return cfg, nil
}

func normalizeAndDeduplicateKeywords(raw []string) []string {
	keywords := make([]string, 0, len(raw))
	exists := make(map[string]struct{})

	for _, keyword := range raw {
		normalized := normalizeSensitiveInput(keyword)
		if normalized == "" {
			continue
		}
		if _, ok := exists[normalized]; ok {
			continue
		}
		exists[normalized] = struct{}{}
		keywords = append(keywords, normalized)
	}

	return keywords
}

func isTargetPackage(path string) bool {
	// Проверяем принадлежность пакета к целевым логгерам.
	return targetPackages[path]
}
