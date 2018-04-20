package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/structtag"
	"github.com/strava/go.strava"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

type TemplateEngine struct {
	templateTypeInfo *ParsedTypeInfo
	cache            map[string]*template.Template
	cacheMutex       sync.Mutex
	sampleActivity   *strava.ActivitySummary
}

type ParsedTypeInfo struct {
	Fields       map[string]string
	StructFields map[string]*ParsedTypeInfo
}

var (
	ActivityTypeInfo  *ParsedTypeInfo
	UserTemplateRegex = regexp.MustCompile("{([^{]*)}")
)

func init() {
	ActivityTypeInfo = parseType(reflect.TypeOf(strava.ActivitySummary{}))
}

func parseType(t reflect.Type) *ParsedTypeInfo {
	timeType := reflect.TypeOf(time.Time{})
	result := &ParsedTypeInfo{Fields: make(map[string]string), StructFields: make(map[string]*ParsedTypeInfo)}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tags, err := structtag.Parse(string(f.Tag))
		if err != nil {
			continue
		}
		jsonTag, err := tags.Get("json")
		if err != nil {
			continue
		}

		result.Fields[jsonTag.Name] = f.Name
		if f.Type.Kind() == reflect.Struct && f.Type != timeType {
			result.StructFields[jsonTag.Name] = parseType(f.Type)
		}
	}
	return result
}

func NewActivityTemplateEngine(appDir string) *TemplateEngine {
	var sampleActivity *strava.ActivitySummary
	content, err := ioutil.ReadFile(filepath.Join(appDir, "samples", "activity.json"))
	if err == nil {
		json.Unmarshal(content, &sampleActivity)
	}

	return &TemplateEngine{
		templateTypeInfo: ActivityTypeInfo,
		cache:            make(map[string]*template.Template),
		sampleActivity:   sampleActivity,
	}
}

func (te *TemplateEngine) Build(userTemplate string) (string, error) {
	result := userTemplate
	matches := UserTemplateRegex.FindAllStringSubmatch(userTemplate, -1)
	for _, match := range matches {
		pattern, err := te.convert(match[1])
		if err != nil {
			return userTemplate, err
		}
		result = strings.Replace(result, match[0], "{{"+pattern+"}}", 1)
	}
	return result, nil
}

func (te *TemplateEngine) convert(userPattern string) (string, error) {
	var partUnit *Unit
	parts := strings.Split(userPattern, ".")
	result := userPattern

	hasError := false
	isPace := false
	currentTypeInfo := te.templateTypeInfo
	for i, part := range parts {
		if name, ok := currentTypeInfo.Fields[part]; ok {
			parts[i] = name
			if typeInfo, ok := currentTypeInfo.StructFields[part]; ok {
				currentTypeInfo = typeInfo
			}
		} else if i == len(parts)-1 {
			isPace = strings.Contains(part, "_pace_") || strings.HasSuffix(part, "_pace")
			if isPace {
				if strings.Contains(part, "_pace_") {
					part = strings.Replace(part, "_pace_", "_speed_", -1)
				} else {
					part = part[:len(part)-len("_pace")] + "_speed_m"
				}
			}

			sepIndex := strings.LastIndex(part, "_")
			if sepIndex >= 0 {
				unitStr := part[sepIndex+1:]
				unit := GetKnownUnit(unitStr)
				if unit != nil {
					if name, ok := currentTypeInfo.Fields[part[:sepIndex]]; ok {
						parts[i] = name
						partUnit = unit
					} else {
						hasError = true
					}
				} else {
					hasError = true
				}
			} else {
				hasError = true
			}
		} else {
			hasError = true
			break
		}
	}
	if hasError {
		return userPattern, fmt.Errorf("invalid field {%s}", userPattern)
	}

	result = "." + strings.Join(parts, ".")
	if strings.HasSuffix(result, "Speed") {
		partUnitName := "m"
		if partUnit != nil {
			partUnitName = partUnit.Name
		}
		funcName := "toSpeed"
		if isPace {
			funcName = "toPace"
		}
		result = fmt.Sprintf("%s %s \"%s\"", funcName, result, partUnitName)
	} else if partUnit != nil {
		result = fmt.Sprintf("toUnit %s \"%s\"", result, partUnit.Name)
	} else if strings.HasSuffix(result, "Time") {
		result = fmt.Sprintf("toTime %s", result)
	}
	return result, nil
}

func (te *TemplateEngine) Compile(userTemplate string) (*template.Template, error) {
	tmpl := te.getTemplateFromCache(userTemplate)
	if tmpl != nil {
		return tmpl, nil
	}

	realTemplate, err := te.Build(userTemplate)
	if err != nil {
		return nil, err
	}
	tmpl, err = template.New(userTemplate).Funcs(template.FuncMap{
		"toUnit":  ConvertToUnit,
		"toSpeed": ConvertToSpeed,
		"toPace":  ConvertToPace,
		"toTime":  getDurationText,
	}).Parse(realTemplate)
	if err != nil {
		return nil, err
	}

	te.putTemplateFromCache(userTemplate, tmpl)
	return tmpl, nil
}

func (te *TemplateEngine) SampleText(tmpl *template.Template) string {
	if te.sampleActivity == nil {
		return ""
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, te.sampleActivity)
	return buf.String()
}

func (te *TemplateEngine) getTemplateFromCache(userTemplate string) *template.Template {
	te.cacheMutex.Lock()
	defer te.cacheMutex.Unlock()
	return te.cache[userTemplate]
}

func (te *TemplateEngine) putTemplateFromCache(userTemplate string, tmpl *template.Template) {
	te.cacheMutex.Lock()
	defer te.cacheMutex.Unlock()
	te.cache[userTemplate] = tmpl
}
