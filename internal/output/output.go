package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Format defines supported output formats.
type Format string

const (
	// FormatJSON renders the unified ok/outcome/data envelope.
	FormatJSON Format = "json"
	// FormatRawJSON renders the raw payload without an envelope.
	FormatRawJSON Format = "raw-json"
	// FormatYAML renders the response as YAML.
	FormatYAML Format = "yaml"
	// FormatTable renders list responses as aligned ASCII columns.
	FormatTable Format = "table"
	// FormatText renders single-line plain-text summaries.
	FormatText Format = "text"
)

// Standard exit codes for Agent & CLI invocation.
const (
	ExitCodeSuccess    = 0
	ExitCodeAPI        = 1
	ExitCodeAuth       = 2
	ExitCodeValidation = 3
	ExitCodeInternal   = 5
)

// Response is the unified envelope designed for Agent / LLM invocation.
type Response struct {
	OK      bool       `json:"ok" yaml:"ok"`
	Outcome string     `json:"outcome" yaml:"outcome"`
	Data    any        `json:"data,omitempty" yaml:"data,omitempty"`
	Meta    *Meta      `json:"meta,omitempty" yaml:"meta,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty" yaml:"error,omitempty"`
}

// Meta describes collection or pagination metadata.
type Meta struct {
	Count int `json:"count,omitempty" yaml:"count,omitempty"`
}

// ErrorInfo describes error details for machine and human consumption.
type ErrorInfo struct {
	Code     int    `json:"code" yaml:"code"`
	Category string `json:"category" yaml:"category"`
	Message  string `json:"message" yaml:"message"`
	Details  any    `json:"details,omitempty" yaml:"details,omitempty"`
}

// Printer handles formatted output to specified writers.
type Printer struct {
	Out    io.Writer
	Err    io.Writer
	Format Format
}

// New creates a Printer with default stdout/stderr and normalized format.
func New(format string) *Printer {
	f := strings.ToLower(strings.TrimSpace(format))
	var targetFormat Format
	switch f {
	case "raw-json", "raw":
		targetFormat = FormatRawJSON
	case "yaml", "yml":
		targetFormat = FormatYAML
	case "table":
		targetFormat = FormatTable
	case "text", "txt":
		targetFormat = FormatText
	default:
		targetFormat = FormatJSON
	}
	return &Printer{
		Format: targetFormat,
		Out:    os.Stdout,
		Err:    os.Stderr,
	}
}

// Success prints a successful response in the configured format.
func (p *Printer) Success(data any) error {
	normalizedData := normalizeData(data)

	switch p.Format {
	case FormatRawJSON:
		return p.printRawJSON(normalizedData)
	case FormatYAML:
		resp := Response{
			OK:      true,
			Outcome: "success",
			Data:    normalizedData,
			Meta:    calculateMeta(normalizedData),
		}
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err := encoder.Encode(resp); err != nil {
			return err
		}
		_, err := fmt.Fprint(p.Out, buf.String())
		return err
	case FormatTable:
		return p.printTable(normalizedData)
	case FormatText:
		return p.printText(normalizedData)
	case FormatJSON:
		fallthrough
	default:
		resp := Response{
			OK:      true,
			Outcome: "success",
			Data:    normalizedData,
			Meta:    calculateMeta(normalizedData),
		}
		encoded, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(p.Out, string(encoded))
		return err
	}
}

// Fail prints an error response in the configured format.
func (p *Printer) Fail(code int, category, message string, details any) {
	errInfo := &ErrorInfo{
		Code:     code,
		Category: category,
		Message:  message,
		Details:  details,
	}
	resp := Response{
		OK:      false,
		Outcome: "failure",
		Error:   errInfo,
	}

	switch p.Format {
	case FormatYAML:
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		_ = encoder.Encode(resp)
		_, _ = fmt.Fprint(p.Err, buf.String())
	case FormatTable, FormatText:
		_, _ = fmt.Fprintf(p.Err, "Error [%s]: %s\n", category, message)
		if details != nil {
			_, _ = fmt.Fprintf(p.Err, "Details: %v\n", details)
		}
	case FormatRawJSON, FormatJSON:
		fallthrough
	default:
		encoded, err := json.MarshalIndent(resp, "", "  ")
		if err == nil {
			_, _ = fmt.Fprintln(p.Err, string(encoded))
		} else {
			_, _ = fmt.Fprintf(p.Err, `{"ok":false,"outcome":"failure","error":{"code":%d,"category":%q,"message":%q}}`+"\n", code, category, message)
		}
	}
}

func calculateMeta(data any) *Meta {
	items, exists := extractCollection(data)
	if exists {
		return &Meta{Count: len(items)}
	}
	return nil
}

func (p *Printer) printRawJSON(data any) error {
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(p.Out, string(encoded))
	return err
}

// printText formats data in readable plain-text lines.
func (p *Printer) printText(data any) error {
	if s, ok := data.(string); ok {
		_, err := fmt.Fprintln(p.Out, s)
		return err
	}

	items, exists := extractCollection(data)
	if exists {
		if len(items) == 0 {
			_, err := fmt.Fprintln(p.Out, "(empty list)")
			return err
		}
		for _, item := range items {
			line := formatItemText(item, "")
			_, _ = fmt.Fprintln(p.Out, line)
		}
		return nil
	}

	// If single map, print key: value
	if m, ok := data.(map[string]any); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := m[k]
			if vMap, isMap := v.(map[string]any); isMap {
				b, _ := json.Marshal(vMap)
				_, _ = fmt.Fprintf(p.Out, "%s: %s\n", k, string(b))
			} else if vSlice, isSlice := v.([]any); isSlice {
				b, _ := json.Marshal(vSlice)
				_, _ = fmt.Fprintf(p.Out, "%s: %s\n", k, string(b))
			} else {
				_, _ = fmt.Fprintf(p.Out, "%s: %v\n", k, v)
			}
		}
		return nil
	}

	return p.printRawJSON(data)
}

// printTable formats data into neat ASCII columns using text/tabwriter.
func (p *Printer) printTable(data any) error {
	items, exists := extractCollection(data)
	if exists {
		return p.renderTableSlice(items, "")
	}

	// If single map, render as 2-column Key-Value table
	if m, ok := data.(map[string]any); ok {
		return p.renderKeyValueTable(m)
	}

	return p.printRawJSON(data)
}

func (p *Printer) renderKeyValueTable(m map[string]any) error {
	if len(m) == 0 {
		_, _ = fmt.Fprintln(p.Out, "(empty)")
		return nil
	}
	w := tabwriter.NewWriter(p.Out, 0, 0, 4, ' ', 0)
	_, _ = fmt.Fprintln(w, "KEY\tVALUE")
	_, _ = fmt.Fprintln(w, "---\t-----")

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		val := m[k]
		valStr := ""
		if val != nil {
			if vMap, ok := val.(map[string]any); ok {
				b, _ := json.Marshal(vMap)
				valStr = string(b)
			} else if vSlice, ok := val.([]any); ok {
				b, _ := json.Marshal(vSlice)
				valStr = string(b)
			} else {
				valStr = fmt.Sprint(val)
			}
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\n", k, valStr)
	}
	return w.Flush()
}

func (p *Printer) renderTableSlice(items []map[string]any, itemType string) error {
	if len(items) == 0 {
		_, _ = fmt.Fprintln(p.Out, "(empty list)")
		return nil
	}

	headers, keys := chooseColumns(items, itemType)

	w := tabwriter.NewWriter(p.Out, 0, 0, 4, ' ', 0)
	_, _ = fmt.Fprintln(w, strings.Join(headers, "\t"))

	seps := make([]string, len(headers))
	for i, h := range headers {
		seps[i] = strings.Repeat("-", maxInt(len(h), 3))
	}
	_, _ = fmt.Fprintln(w, strings.Join(seps, "\t"))

	for _, item := range items {
		row := make([]string, len(keys))
		for i, k := range keys {
			val := item[k]
			if val == nil {
				row[i] = "-"
			} else {
				row[i] = strings.ReplaceAll(fmt.Sprint(val), "\t", " ")
			}
		}
		_, _ = fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	return w.Flush()
}

type entitySpec struct {
	indicator func(m map[string]any) bool
	headers   []string
	keys      []string
}

var knownEntitySpecs = []entitySpec{
	{
		indicator: func(m map[string]any) bool { return m["title"] != nil && m["severity"] != nil },
		headers:   []string{"ID", "TITLE", "STATUS", "SEVERITY", "PRI", "ASSIGNED_TO", "OPENED_BY"},
		keys:      []string{"id", "title", "status", "severity", "pri", "assignedTo", "openedBy"},
	},
	{
		indicator: func(m map[string]any) bool {
			return m["name"] != nil && (m["consumed"] != nil || m["estimate"] != nil) &&
				(m["deadline"] != nil || m["execution"] != nil || m["status"] != nil)
		},
		headers: []string{"ID", "NAME", "STATUS", "PRI", "ASSIGNED_TO", "ESTIMATE", "CONSUMED", "DEADLINE"},
		keys:    []string{"id", "name", "status", "pri", "assignedTo", "estimate", "consumed", "deadline"},
	},
	{
		indicator: func(m map[string]any) bool { return m["name"] != nil && m["date"] != nil && m["begin"] != nil },
		headers:   []string{"ID", "NAME", "STATUS", "DATE", "BEGIN", "END", "PRI", "TYPE"},
		keys:      []string{"id", "name", "status", "date", "begin", "end", "pri", "type"},
	},
	{
		indicator: func(m map[string]any) bool { return m["title"] != nil && m["estimate"] != nil && m["openedBy"] != nil },
		headers:   []string{"ID", "TITLE", "STATUS", "PRI", "ESTIMATE", "ASSIGNED_TO", "OPENED_BY"},
		keys:      []string{"id", "title", "status", "pri", "estimate", "assignedTo", "openedBy"},
	},
	{
		indicator: func(m map[string]any) bool { return m["name"] != nil && m["code"] != nil && m["PM"] != nil },
		headers:   []string{"ID", "NAME", "CODE", "STATUS", "BEGIN", "END", "PM"},
		keys:      []string{"id", "name", "code", "status", "begin", "end", "PM"},
	},
	{
		indicator: func(m map[string]any) bool { return m["name"] != nil && m["code"] != nil && m["PO"] != nil },
		headers:   []string{"ID", "NAME", "CODE", "STATUS", "PO", "QD", "RD"},
		keys:      []string{"id", "name", "code", "status", "PO", "QD", "RD"},
	},
	{
		indicator: func(m map[string]any) bool { return m["account"] != nil && m["realname"] != nil },
		headers:   []string{"ID", "ACCOUNT", "REALNAME", "ROLE", "EMAIL", "GENDER", "MOBILE"},
		keys:      []string{"id", "account", "realname", "role", "email", "gender", "mobile"},
	},
	{
		indicator: func(m map[string]any) bool { return m["name"] != nil && m["parent"] != nil && m["manager"] != nil },
		headers:   []string{"ID", "NAME", "PARENT", "MANAGER"},
		keys:      []string{"id", "name", "parent", "manager"},
	},
	{
		indicator: func(m map[string]any) bool { return m["actor"] != nil && m["action"] != nil && m["objectType"] != nil },
		headers:   []string{"ID", "DATE", "ACTOR", "ACTION", "OBJECT_TYPE", "OBJECT_NAME"},
		keys:      []string{"id", "date", "actor", "action", "objectType", "objectName"},
	},
}

func chooseColumns(items []map[string]any, itemType string) ([]string, []string) {
	if len(items) == 0 {
		return nil, nil
	}
	sample := items[0]
	for _, spec := range knownEntitySpecs {
		if spec.indicator(sample) {
			return spec.headers, spec.keys
		}
	}
	return fallbackColumns(items)
}

func fallbackColumns(items []map[string]any) ([]string, []string) {
	keySet := make(map[string]struct{})
	for _, item := range items {
		for k := range item {
			keySet[k] = struct{}{}
		}
	}
	var sortedKeys []string
	for k := range keySet {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	headers := make([]string, len(sortedKeys))
	for i, k := range sortedKeys {
		headers[i] = strings.ToUpper(k)
	}
	return headers, sortedKeys
}

func formatItemText(item map[string]any, itemType string) string {
	id := fmt.Sprint(item["id"])
	if item["title"] != nil && item["severity"] != nil {
		return fmt.Sprintf("#%s %s [%s] (severity: %v, pri: %v, assigned: %v)",
			id, item["title"], item["status"], item["severity"], item["pri"], item["assignedTo"])
	}
	if item["name"] != nil && (item["consumed"] != nil || item["estimate"] != nil) {
		return fmt.Sprintf("#%s %s [%s] (pri: %v, est: %vh, assigned: %v)",
			id, item["name"], item["status"], item["pri"], item["estimate"], item["assignedTo"])
	}
	if item["name"] != nil && item["date"] != nil && item["begin"] != nil {
		return fmt.Sprintf("#%s %s [%s] (date: %v, time: %v~%v, pri: %v)",
			id, item["name"], item["status"], item["date"], item["begin"], item["end"], item["pri"])
	}
	if item["title"] != nil && item["estimate"] != nil {
		return fmt.Sprintf("#%s %s [%s] (pri: %v, est: %vh, assigned: %v)",
			id, item["title"], item["status"], item["pri"], item["estimate"], item["assignedTo"])
	}
	if item["name"] != nil && item["code"] != nil && item["PM"] != nil {
		return fmt.Sprintf("#%s %s (%s) [%s] (PM: %v, %v ~ %v)",
			id, item["name"], item["code"], item["status"], item["PM"], item["begin"], item["end"])
	}
	if item["name"] != nil && item["code"] != nil && item["PO"] != nil {
		return fmt.Sprintf("#%s %s (%s) [%s] (PO: %v)",
			id, item["name"], item["code"], item["status"], item["PO"])
	}
	if item["account"] != nil && item["realname"] != nil {
		return fmt.Sprintf("#%s %s (%s) [%s] email: %v",
			id, item["account"], item["realname"], item["role"], item["email"])
	}

	name := item["name"]
	if name == nil {
		name = item["title"]
	}
	return fmt.Sprintf("#%s %v [%v]", id, name, item["status"])
}

func extractCollection(data any) ([]map[string]any, bool) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Slice {
		if slice, ok := toMapSlice(data); ok {
			return slice, true
		}
	}

	if m, ok := data.(map[string]any); ok {
		candidateKeys := []string{"projectStats", "projects", "todos", "stories", "bugs", "tasks", "products", "users", "depts", "dynamics", "actions", "sons", "tree", "items"}
		for _, k := range candidateKeys {
			if v, exists := m[k]; exists {
				// Special check: in product/all, products is map[string]string {"1": "Name1"}
				if strMap, isStrMap := v.(map[string]any); isStrMap && k == "products" {
					var converted []map[string]any
					for prodID, prodName := range strMap {
						if _, isObj := prodName.(map[string]any); !isObj {
							converted = append(converted, map[string]any{
								"id":   prodID,
								"name": fmt.Sprint(prodName),
							})
						}
					}
					if len(converted) > 0 {
						sort.Slice(converted, func(i, j int) bool {
							return fmt.Sprint(converted[i]["id"]) < fmt.Sprint(converted[j]["id"])
						})
						return converted, true
					}
				}

				if slice, ok := toMapSlice(v); ok {
					return slice, true
				}
			}
		}
	}

	return nil, false
}

func toMapSlice(v any) ([]map[string]any, bool) {
	if v == nil {
		return nil, false
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Slice {
		var out []map[string]any
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i).Interface()
			if elemMap, ok := elem.(map[string]any); ok {
				out = append(out, elemMap)
			}
		}
		return out, true
	}

	if val.Kind() == reflect.Map {
		var out []map[string]any
		keys := val.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
		})
		for _, k := range keys {
			elem := val.MapIndex(k).Interface()
			if elemMap, ok := elem.(map[string]any); ok {
				out = append(out, elemMap)
			}
		}
		if len(out) > 0 {
			return out, true
		}
	}

	return nil, false
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func normalizeData(data any) any {
	if raw, ok := data.(json.RawMessage); ok {
		var parsed any
		if err := json.Unmarshal(raw, &parsed); err == nil {
			return parsed
		}
		return string(raw)
	}
	return data
}
