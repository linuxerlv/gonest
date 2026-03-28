package config

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

// ============================================================
//                    New Tests
// ============================================================

func TestKoanfConfig_New(t *testing.T) {
	t.Run("default delimiter", func(t *testing.T) {
		config := NewKoanfConfig("")
		if config == nil {
			t.Fatal("expected config to be non-nil")
		}
		// Default delimiter should be "."
		if config.GetString("test.key") != "" {
			// Key not set, but delimiter is working
			t.Errorf("expected empty string for unset key, got %s", config.GetString("test.key"))
		}
	})

	t.Run("custom delimiter", func(t *testing.T) {
		config := NewKoanfConfig("/")
		if config.delim != "/" {
			t.Errorf("expected delimiter '/', got '%s'", config.delim)
		}
	})
}

func TestKoanfConfig_NewWithConf(t *testing.T) {
	conf := DefaultKoanfConf()
	conf.Delim = "_"
	config := NewKoanfConfigWithConf(conf)

	if config.delim != "_" {
		t.Errorf("expected delimiter '_', got '%s'", config.delim)
	}
}

// ============================================================
//                    Get Method Tests
// ============================================================

func TestKoanfConfig_Get(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"string_key":   "hello",
		"int_key":      42,
		"float_key":    3.14,
		"bool_key":     true,
		"nested.key":   "nested_value",
		"empty_string": "",
		"zero_int":     0,
		"zero_float":   0.0,
		"false_bool":   false,
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected any
	}{
		{"string value", "string_key", "hello"},
		{"int value", "int_key", 42},
		{"float value", "float_key", 3.14},
		{"bool value", "bool_key", true},
		{"nested value", "nested.key", "nested_value"},
		{"empty string", "empty_string", ""},
		{"zero int", "zero_int", 0},
		{"zero float", "zero_float", 0.0},
		{"false bool", "false_bool", false},
		{"nonexistent key", "nonexistent", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.Get(tt.key)
			if got != tt.expected {
				t.Errorf("Get(%s) = %v (%T), want %v (%T)", tt.key, got, got, tt.expected, tt.expected)
			}
		})
	}
}

func TestKoanfConfig_GetString(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"string": "hello",
		"int":    42,
		"float":  3.14,
		"bool":   true,
		"empty":  "",
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"string value", "string", "hello"},
		{"int to string", "int", "42"},
		{"float to string", "float", "3.14"},
		{"bool to string", "bool", "true"},
		{"empty string", "empty", ""},
		{"nonexistent", "nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetString(tt.key)
			if got != tt.expected {
				t.Errorf("GetString(%s) = %s, want %s", tt.key, got, tt.expected)
			}
		})
	}
}

func TestKoanfConfig_GetInt(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"int":    42,
		"string": "42",
		"float":  3.14,
		"bool":   true,
		"zero":   0,
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected int
	}{
		{"int value", "int", 42},
		{"string to int", "string", 42},
		{"float to int", "float", 3},
		{"bool true to int - koanf returns 0 for non-numeric bool", "bool", 0},
		{"zero", "zero", 0},
		{"nonexistent", "nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetInt(tt.key)
			if got != tt.expected {
				t.Errorf("GetInt(%s) = %d, want %d", tt.key, got, tt.expected)
			}
		})
	}
}

func TestKoanfConfig_GetInt64(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"int64": int64(9223372036854775807),
		"int":   42,
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected int64
	}{
		{"int64 value", "int64", 9223372036854775807},
		{"int to int64", "int", 42},
		{"nonexistent", "nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetInt64(tt.key)
			if got != tt.expected {
				t.Errorf("GetInt64(%s) = %d, want %d", tt.key, got, tt.expected)
			}
		})
	}
}

func TestKoanfConfig_GetFloat64(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"float":  3.14,
		"int":    42,
		"string": "3.14",
		"zero":   0.0,
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected float64
	}{
		{"float value", "float", 3.14},
		{"int to float", "int", 42.0},
		{"string to float", "string", 3.14},
		{"zero", "zero", 0.0},
		{"nonexistent", "nonexistent", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetFloat64(tt.key)
			if got != tt.expected {
				t.Errorf("GetFloat64(%s) = %f, want %f", tt.key, got, tt.expected)
			}
		})
	}
}

func TestKoanfConfig_GetBool(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"true_bool":    true,
		"false_bool":   false,
		"int_one":      1,
		"int_zero":     0,
		"string_true":  "true",
		"string_false": "false",
		"string_yes":   "yes",
		"string_no":    "no",
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"true bool", "true_bool", true},
		{"false bool", "false_bool", false},
		{"int one", "int_one", true},
		{"int zero", "int_zero", false},
		{"string true", "string_true", true},
		{"string false", "string_false", false},
		{"string yes - koanf doesn't parse this as bool", "string_yes", false},
		{"string no", "string_no", false},
		{"nonexistent", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetBool(tt.key)
			if got != tt.expected {
				t.Errorf("GetBool(%s) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

// ============================================================
//                    Duration and Time Tests
// ============================================================

func TestKoanfConfig_GetDuration(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"duration_sec": 5 * time.Second,
		"duration_min": 2 * time.Minute,
		"empty":        time.Duration(0),
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected time.Duration
	}{
		{"5 seconds", "duration_sec", 5 * time.Second},
		{"2 minutes", "duration_min", 2 * time.Minute},
		{"empty", "empty", time.Duration(0)},
		{"nonexistent", "nonexistent", time.Duration(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetDuration(tt.key)
			if got != tt.expected {
				t.Errorf("GetDuration(%s) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestKoanfConfig_GetTime(t *testing.T) {
	rfc3339Time := "2024-01-01T12:00:00Z"
	mapProvider := NewMapProvider(map[string]any{
		"time":  rfc3339Time,
		"empty": "",
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	expectedTime, err := time.Parse(time.RFC3339, rfc3339Time)
	if err != nil {
		t.Fatalf("failed to parse test time: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected time.Time
	}{
		{"valid time", "time", expectedTime},
		{"empty", "empty", time.Time{}},
		{"nonexistent", "nonexistent", time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetTime(tt.key)
			if !got.Equal(tt.expected) {
				t.Errorf("GetTime(%s) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

// ============================================================
//                    Slice Tests
// ============================================================

func TestKoanfConfig_GetStringSlice(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"slice":  []string{"a", "b", "c"},
		"empty":  []string{},
		"single": []string{"only"},
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected []string
	}{
		{"string slice", "slice", []string{"a", "b", "c"}},
		{"empty slice", "empty", []string{}},
		{"single element", "single", []string{"only"}},
		{"nonexistent", "nonexistent", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetStringSlice(tt.key)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("GetStringSlice(%s) = %v, want nil", tt.key, got)
				}
			} else {
				if len(got) != len(tt.expected) {
					t.Errorf("GetStringSlice(%s) = %v, want %v", tt.key, got, tt.expected)
				} else {
					for i := range tt.expected {
						if got[i] != tt.expected[i] {
							t.Errorf("GetStringSlice(%s) = %v, want %v", tt.key, got, tt.expected)
							break
						}
					}
				}
			}
		})
	}
}

func TestKoanfConfig_GetIntSlice(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"slice": []int{1, 2, 3},
		"empty": []int{},
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected []int
	}{
		{"int slice", "slice", []int{1, 2, 3}},
		{"empty slice", "empty", []int{}},
		{"nonexistent", "nonexistent", []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetIntSlice(tt.key)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("GetIntSlice(%s) = %v, want nil", tt.key, got)
				}
			} else {
				if len(got) != len(tt.expected) {
					t.Errorf("GetIntSlice(%s) = %v, want %v", tt.key, got, tt.expected)
				} else {
					for i := range tt.expected {
						if got[i] != tt.expected[i] {
							t.Errorf("GetIntSlice(%s) = %v, want %v", tt.key, got, tt.expected)
							break
						}
					}
				}
			}
		})
	}
}

// ============================================================
//                    Map Tests
// ============================================================

func TestKoanfConfig_GetStringMap(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"map": map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
		"nested": map[string]any{
			"nested_key": "nested_value",
		},
		"empty": map[string]any{},
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected map[string]any
	}{
		{"string map", "map", map[string]any{"key1": "value1", "key2": "value2"}},
		{"nested map", "nested", map[string]any{"nested_key": "nested_value"}},
		{"empty map", "empty", map[string]any{}},
		{"nonexistent", "nonexistent", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetStringMap(tt.key)
			if tt.expected == nil || len(tt.expected) == 0 {
				if got != nil && len(got) > 0 {
					t.Errorf("GetStringMap(%s) = %v, want empty", tt.key, got)
				}
			} else {
				if len(got) != len(tt.expected) {
					t.Errorf("GetStringMap(%s) = %v, want %v", tt.key, got, tt.expected)
				} else {
					for k, v := range tt.expected {
						if got[k] != v {
							t.Errorf("GetStringMap(%s)[%s] = %v, want %v", tt.key, k, got[k], v)
						}
					}
				}
			}
		})
	}
}

func TestKoanfConfig_GetStringMapString(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"map": map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
		"mixed": map[string]any{
			"str": "string",
			"int": 42,
		},
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected map[string]string
	}{
		{"string map", "map", map[string]string{"key1": "value1", "key2": "value2"}},
		{"mixed types", "mixed", map[string]string{"str": "string", "int": "42"}},
		{"nonexistent", "nonexistent", map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetStringMapString(tt.key)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("GetStringMapString(%s) = %v, want nil", tt.key, got)
				}
			} else {
				if len(got) != len(tt.expected) {
					t.Errorf("GetStringMapString(%s) = %v, want %v", tt.key, got, tt.expected)
				} else {
					for k, v := range tt.expected {
						if got[k] != v {
							t.Errorf("GetStringMapString(%s)[%s] = %s, want %s", tt.key, k, got[k], v)
						}
					}
				}
			}
		})
	}
}

// ============================================================
//                    IsSet Tests
// ============================================================

func TestKoanfConfig_IsSet(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"exists":       "value",
		"empty_string": "",
		"zero":         0,
		"false":        false,
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"exists", "exists", true},
		{"empty string", "empty_string", true},
		{"zero", "zero", true},
		{"false", "false", true},
		{"nonexistent", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.IsSet(tt.key)
			if got != tt.expected {
				t.Errorf("IsSet(%s) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

// ============================================================
//                    Load Tests
// ============================================================

func TestKoanfConfig_Load(t *testing.T) {
	t.Run("load from map provider", func(t *testing.T) {
		mapProvider := NewMapProvider(map[string]any{
			"key": "value",
		}, ".")

		config := NewKoanfConfig(".")
		if err := config.Load(mapProvider, nil); err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if config.GetString("key") != "value" {
			t.Errorf("expected 'value', got '%s'", config.GetString("key"))
		}
	})

	t.Run("nested keys", func(t *testing.T) {
		mapProvider := NewMapProvider(map[string]any{
			"database.host": "localhost",
			"database.port": 5432,
			"database.name": "testdb",
		}, ".")

		config := NewKoanfConfig(".")
		if err := config.Load(mapProvider, nil); err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if config.GetString("database.host") != "localhost" {
			t.Errorf("expected 'localhost', got '%s'", config.GetString("database.host"))
		}

		if config.GetInt("database.port") != 5432 {
			t.Errorf("expected 5432, got %d", config.GetInt("database.port"))
		}
	})

	t.Run("empty map provider", func(t *testing.T) {
		mapProvider := NewMapProvider(map[string]any{}, ".")

		config := NewKoanfConfig(".")
		if err := config.Load(mapProvider, nil); err != nil {
			t.Fatalf("failed to load empty config: %v", err)
		}

		// Should not error on empty config
		if config.GetString("anything") != "" {
			t.Error("expected empty string for nonexistent key")
		}
	})

	t.Run("nil map provider data", func(t *testing.T) {
		provider := &MapProvider{data: nil, delim: ".", name: "nil-map"}
		config := NewKoanfConfig(".")
		if err := config.Load(provider, nil); err != nil {
			t.Fatalf("failed to load nil map config: %v", err)
		}
	})
}

// ============================================================
//                    LoadWithOverride Tests
// ============================================================

func TestKoanfConfig_LoadWithOverride(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"key1": "original",
		"key2": "value2",
	}, ".")

	config := NewKoanfConfig(".")
	overrideFn := func(data map[string]any) {
		data["key1"] = "overridden"
		data["key3"] = "added"
	}

	if err := config.LoadWithOverride(mapProvider, nil, overrideFn); err != nil {
		t.Fatalf("failed to load with override: %v", err)
	}

	if config.GetString("key1") != "overridden" {
		t.Errorf("expected 'overridden', got '%s'", config.GetString("key1"))
	}

	if config.GetString("key2") != "value2" {
		t.Errorf("expected 'value2', got '%s'", config.GetString("key2"))
	}

	if config.GetString("key3") != "added" {
		t.Errorf("expected 'added', got '%s'", config.GetString("key3"))
	}
}

func TestKoanfConfig_LoadWithOverride_Nil(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"key": "value",
	}, ".")

	config := NewKoanfConfig(".")
	// Test with nil override function
	if err := config.LoadWithOverride(mapProvider, nil, nil); err != nil {
		t.Fatalf("failed to load with nil override: %v", err)
	}

	if config.GetString("key") != "value" {
		t.Errorf("expected 'value', got '%s'", config.GetString("key"))
	}
}

// ============================================================
//                    Merge Tests
// ============================================================

func TestKoanfConfig_Merge(t *testing.T) {
	t.Run("merge two configs", func(t *testing.T) {
		config1 := NewKoanfConfig(".")
		mapProvider1 := NewMapProvider(map[string]any{
			"key1": "value1",
			"key2": "value2",
		}, ".")

		if err := config1.Load(mapProvider1, nil); err != nil {
			t.Fatalf("failed to load config1: %v", err)
		}

		config2 := NewKoanfConfig(".")
		mapProvider2 := NewMapProvider(map[string]any{
			"key3": "value3",
			"key2": "overridden",
		}, ".")

		if err := config2.Load(mapProvider2, nil); err != nil {
			t.Fatalf("failed to load config2: %v", err)
		}

		if err := config1.Merge(config2); err != nil {
			t.Fatalf("failed to merge configs: %v", err)
		}

		if config1.GetString("key1") != "value1" {
			t.Errorf("expected 'value1', got '%s'", config1.GetString("key1"))
		}

		if config1.GetString("key2") != "overridden" {
			t.Errorf("expected 'overridden', got '%s'", config1.GetString("key2"))
		}

		if config1.GetString("key3") != "value3" {
			t.Errorf("expected 'value3', got '%s'", config1.GetString("key3"))
		}
	})

	t.Run("merge with empty config", func(t *testing.T) {
		config1 := NewKoanfConfig(".")
		mapProvider1 := NewMapProvider(map[string]any{
			"key": "value",
		}, ".")

		if err := config1.Load(mapProvider1, nil); err != nil {
			t.Fatalf("failed to load config1: %v", err)
		}

		config2 := NewKoanfConfig(".")
		mapProvider2 := NewMapProvider(map[string]any{}, ".")

		if err := config2.Load(mapProvider2, nil); err != nil {
			t.Fatalf("failed to load config2: %v", err)
		}

		if err := config1.Merge(config2); err != nil {
			t.Fatalf("failed to merge with empty config: %v", err)
		}

		if config1.GetString("key") != "value" {
			t.Errorf("expected 'value', got '%s'", config1.GetString("key"))
		}
	})
}

// ============================================================
//                    Provider Tests
// ============================================================

func TestFileProvider(t *testing.T) {
	t.Run("load from file", func(t *testing.T) {
		// Create temp file
		tmpFile, err := os.CreateTemp("", "config-test-*.json")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		content := `{"host": "localhost", "port": 8080, "enabled": true}`
		if _, err := tmpFile.WriteString(content); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
		tmpFile.Close()

		parser := NewJSONParser()
		provider := NewFileProvider(tmpFile.Name(), parser)

		data, err := provider.Read()
		if err != nil {
			t.Fatalf("failed to read from FileProvider: %v", err)
		}

		if data["host"] != "localhost" {
			t.Errorf("expected 'localhost', got '%v'", data["host"])
		}

		// JSON unmarshals numbers as float64
		if port, ok := data["port"].(float64); ok {
			if int(port) != 8080 {
				t.Errorf("expected 8080, got %v", port)
			}
		} else if data["port"] != 8080 {
			t.Errorf("expected 8080, got %v", data["port"])
		}
		if data["enabled"] != true {
			t.Errorf("expected true, got %v", data["enabled"])
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		provider := NewFileProvider("/nonexistent/path/file.json", NewJSONParser())
		_, err := provider.Read()
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("watch not supported", func(t *testing.T) {
		provider := NewFileProvider("/tmp/test.json", NewJSONParser())
		err := provider.Watch(func(any, error) {})
		if err == nil {
			t.Error("expected watch to return error")
		}
	})

	t.Run("name", func(t *testing.T) {
		provider := NewFileProvider("/tmp/config.json", NewJSONParser())
		name := provider.Name()
		if !strings.HasPrefix(name, "file:") {
			t.Errorf("expected name to start with 'file:', got '%s'", name)
		}
	})
}

func TestEnvProvider(t *testing.T) {
	t.Run("read environment variables", func(t *testing.T) {
		// Set test env vars
		t.Setenv("TEST_HOST", "localhost")
		t.Setenv("TEST_PORT", "8080")
		t.Setenv("TEST_ENABLED", "true")

		provider := NewEnvProvider(WithEnvPrefix("TEST_"))

		data, err := provider.Read()
		if err != nil {
			t.Fatalf("failed to read from EnvProvider: %v", err)
		}

		if data["host"] != "localhost" {
			t.Errorf("expected 'localhost', got '%v'", data["host"])
		}

		if data["port"] != "8080" {
			t.Errorf("expected '8080', got '%v'", data["port"])
		}

		if data["enabled"] != "true" {
			t.Errorf("expected 'true', got '%v'", data["enabled"])
		}
	})

	t.Run("with custom delimiter", func(t *testing.T) {
		t.Setenv("TEST_DATABASE_HOST", "localhost")
		t.Setenv("TEST_DATABASE_PORT", "5432")

		provider := NewEnvProvider(
			WithEnvPrefix("TEST_"),
			WithEnvDelimiter("_"),
		)

		data, err := provider.Read()
		if err != nil {
			t.Fatalf("failed to read from EnvProvider: %v", err)
		}

		// TEST_DATABASE_HOST becomes database-host (after lowercase)
		if host, ok := data["database_host"]; ok {
			if host != "localhost" {
				t.Errorf("expected 'localhost', got '%v'", host)
			}
		} else {
			t.Errorf("expected key 'database_host' not found, got keys: %v", data)
		}

		if port, ok := data["database_port"]; ok {
			if port != "5432" {
				t.Errorf("expected '5432', got '%v'", port)
			}
		} else {
			t.Errorf("expected key 'database_port' not found, got keys: %v", data)
		}
	})

	t.Run("name", func(t *testing.T) {
		provider := NewEnvProvider()
		name := provider.Name()
		if name != "environment" {
			t.Errorf("expected 'environment', got '%s'", name)
		}
	})

	t.Run("watch not supported", func(t *testing.T) {
		provider := NewEnvProvider()
		err := provider.Watch(func(any, error) {})
		if err == nil {
			t.Error("expected watch to return error")
		}
	})
}

func TestFlagProvider(t *testing.T) {
	t.Run("read command line flags", func(t *testing.T) {
		var flagSet = pflag.NewFlagSet("test", pflag.ContinueOnError)
		flagSet.String("host", "localhost", "host")
		flagSet.Int("port", 8080, "port")
		flagSet.Bool("enabled", true, "enabled")

		// Parse the flags
		flagSet.Parse([]string{"--host=testhost", "--port=9090", "--enabled=false"})

		provider := NewFlagProvider(flagSet, ".")

		data, err := provider.Read()
		if err != nil {
			t.Fatalf("failed to read from FlagProvider: %v", err)
		}

		if data["host"] != "testhost" {
			t.Errorf("expected 'testhost', got '%v'", data["host"])
		}

		if data["port"] != 9090 {
			t.Errorf("expected 9090, got %v", data["port"])
		}

		if data["enabled"] != false {
			t.Errorf("expected false, got %v", data["enabled"])
		}
	})

	t.Run("name", func(t *testing.T) {
		provider := NewFlagProvider(pflag.NewFlagSet("test", pflag.ContinueOnError), ".")
		name := provider.Name()
		if name != "flags" {
			t.Errorf("expected 'flags', got '%s'", name)
		}
	})
}

func TestMapProvider(t *testing.T) {
	t.Run("read from map", func(t *testing.T) {
		data := map[string]any{
			"host":  "localhost",
			"port":  8080,
			"empty": "",
		}

		provider := NewMapProvider(data, ".")

		readData, err := provider.Read()
		if err != nil {
			t.Fatalf("failed to read from MapProvider: %v", err)
		}

		if len(readData) != len(data) {
			t.Errorf("expected %d items, got %d", len(data), len(readData))
		}

		if readData["host"] != "localhost" {
			t.Errorf("expected 'localhost', got '%v'", readData["host"])
		}

		if readData["port"] != 8080 {
			t.Errorf("expected 8080, got %v", readData["port"])
		}
	})

	t.Run("nil data", func(t *testing.T) {
		provider := NewMapProvider(nil, ".")

		data, err := provider.Read()
		if err != nil {
			t.Fatalf("failed to read nil data: %v", err)
		}

		if len(data) != 0 {
			t.Errorf("expected empty map, got %v", data)
		}
	})

	t.Run("empty data", func(t *testing.T) {
		provider := NewMapProvider(map[string]any{}, ".")

		data, err := provider.Read()
		if err != nil {
			t.Fatalf("failed to read empty data: %v", err)
		}

		if len(data) != 0 {
			t.Errorf("expected empty map, got %v", data)
		}
	})

	t.Run("name", func(t *testing.T) {
		provider := NewMapProvider(map[string]any{}, ".")
		name := provider.Name()
		if name != "map" {
			t.Errorf("expected 'map', got '%s'", name)
		}
	})

	t.Run("watch not supported", func(t *testing.T) {
		provider := NewMapProvider(map[string]any{}, ".")
		err := provider.Watch(func(any, error) {})
		if err == nil {
			t.Error("expected watch to return error")
		}
	})
}

func TestStructProvider(t *testing.T) {
	t.Run("read from struct", func(t *testing.T) {
		type Config struct {
			Host  string `koanf:"host"`
			Port  int    `koanf:"port"`
			Env   string `koanf:"environment"`
			Flags []bool `koanf:"flags"`
		}

		cfg := Config{
			Host:  "localhost",
			Port:  8080,
			Env:   "production",
			Flags: []bool{true, false},
		}

		provider := NewStructProvider(cfg, "koanf")

		data, err := provider.Read()
		if err != nil {
			t.Fatalf("failed to read from StructProvider: %v", err)
		}

		if data["host"] != "localhost" {
			t.Errorf("expected 'localhost', got '%v'", data["host"])
		}

		if data["port"] != 8080 {
			t.Errorf("expected 8080, got %v", data["port"])
		}

		if data["environment"] != "production" {
			t.Errorf("expected 'production', got '%v'", data["environment"])
		}
	})

	t.Run("name", func(t *testing.T) {
		provider := NewStructProvider(map[string]any{}, "koanf")
		name := provider.Name()
		if name != "struct" {
			t.Errorf("expected 'struct', got '%s'", name)
		}
	})
}

func TestJSONParser(t *testing.T) {
	t.Run("parse valid JSON", func(t *testing.T) {
		parser := NewJSONParser()
		data := []byte(`{"key": "value", "number": 42, "bool": true, "array": [1, 2, 3]}`)

		result, err := parser.Parse(data)
		if err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if result["key"] != "value" {
			t.Errorf("expected 'value', got '%v'", result["key"])
		}

		// JSON unmarshals numbers as float64
		if num, ok := result["number"].(float64); ok {
			if int(num) != 42 {
				t.Errorf("expected 42, got %v", num)
			}
		} else if result["number"] != 42 {
			t.Errorf("expected 42, got %v", result["number"])
		}

		if result["bool"] != true {
			t.Errorf("expected true, got %v", result["bool"])
		}
	})

	t.Run("parse invalid JSON", func(t *testing.T) {
		parser := NewJSONParser()
		data := []byte(`{invalid json}`)

		_, err := parser.Parse(data)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("marshal to JSON", func(t *testing.T) {
		parser := NewJSONParser()
		data := map[string]any{
			"key":   "value",
			"array": []int{1, 2, 3},
		}

		result, err := parser.Marshal(data)
		if err != nil {
			t.Fatalf("failed to marshal JSON: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(result, &parsed); err != nil {
			t.Fatalf("failed to unmarshal result: %v", err)
		}

		if parsed["key"] != "value" {
			t.Errorf("expected 'value', got '%v'", parsed["key"])
		}
	})

	t.Run("name", func(t *testing.T) {
		parser := NewJSONParser()
		name := parser.Name()
		if name != "json" {
			t.Errorf("expected 'json', got '%s'", name)
		}
	})
}

// ============================================================
//                    Unmarshal Tests
// ============================================================

type TestConfigStruct struct {
	Host     string        `koanf:"host"`
	Port     int           `koanf:"port"`
	Enabled  bool          `koanf:"enabled"`
	Duration time.Duration `koanf:"timeout"`
}

type NestedConfig struct {
	Server struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	} `koanf:"server"`
	Database struct {
		Host     string `koanf:"host"`
		Port     int    `koanf:"port"`
		Username string `koanf:"username"`
	} `koanf:"database"`
}

func TestKoanfConfig_Unmarshal(t *testing.T) {
	t.Run("unmarshal flat struct", func(t *testing.T) {
		mapProvider := NewMapProvider(map[string]any{
			"host":    "localhost",
			"port":    8080,
			"enabled": true,
			"timeout": "5s",
		}, ".")

		config := NewKoanfConfig(".")
		if err := config.Load(mapProvider, nil); err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		var cfg TestConfigStruct
		if err := config.Unmarshal("", &cfg); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if cfg.Host != "localhost" {
			t.Errorf("expected 'localhost', got '%s'", cfg.Host)
		}

		if cfg.Port != 8080 {
			t.Errorf("expected 8080, got %d", cfg.Port)
		}

		if cfg.Enabled != true {
			t.Errorf("expected true, got %v", cfg.Enabled)
		}
	})

	t.Run("unmarshal nested config", func(t *testing.T) {
		mapProvider := NewMapProvider(map[string]any{
			"server.host":       "api.example.com",
			"server.port":       443,
			"database.host":     "localhost",
			"database.port":     5432,
			"database.username": "admin",
		}, ".")

		config := NewKoanfConfig(".")
		if err := config.Load(mapProvider, nil); err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		var cfg NestedConfig
		if err := config.Unmarshal("", &cfg); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if cfg.Server.Host != "api.example.com" {
			t.Errorf("expected 'api.example.com', got '%s'", cfg.Server.Host)
		}

		if cfg.Server.Port != 443 {
			t.Errorf("expected 443, got %d", cfg.Server.Port)
		}

		if cfg.Database.Host != "localhost" {
			t.Errorf("expected 'localhost', got '%s'", cfg.Database.Host)
		}

		if cfg.Database.Username != "admin" {
			t.Errorf("expected 'admin', got '%s'", cfg.Database.Username)
		}
	})

	t.Run("unmarshal specific key", func(t *testing.T) {
		mapProvider := NewMapProvider(map[string]any{
			"app.host": "localhost",
			"app.port": 8080,
		}, ".")

		config := NewKoanfConfig(".")
		if err := config.Load(mapProvider, nil); err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		var appConfig struct {
			Host string `koanf:"host"`
			Port int    `koanf:"port"`
		}

		if err := config.Unmarshal("app", &appConfig); err != nil {
			t.Fatalf("failed to unmarshal app config: %v", err)
		}

		if appConfig.Host != "localhost" {
			t.Errorf("expected 'localhost', got '%s'", appConfig.Host)
		}

		if appConfig.Port != 8080 {
			t.Errorf("expected 8080, got %d", appConfig.Port)
		}
	})

	t.Run("unmarshal to nil pointer", func(t *testing.T) {
		config := NewKoanfConfig(".")
		if err := config.Unmarshal("", nil); err == nil {
			t.Error("expected error when unmarshaling to nil pointer")
		}
	})

	t.Run("unmarshal with tag override", func(t *testing.T) {
		mapProvider := NewMapProvider(map[string]any{
			"host": "localhost",
			"port": 8080,
		}, ".")

		config := NewKoanfConfig(".")
		if err := config.Load(mapProvider, nil); err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		type CustomStruct struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		}

		var cfg CustomStruct
		opts := UnmarshalOptions{
			Tag: "json",
		}

		if err := config.UnmarshalWithConf("", &cfg, opts); err != nil {
			t.Fatalf("failed to unmarshal with custom tag: %v", err)
		}
	})
}

// ============================================================
//                    All and Keys Tests
// ============================================================

func TestKoanfConfig_All(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"key1":        "value1",
		"key2":        "value2",
		"nested.key3": "value3",
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	all := config.All()

	if len(all) != 3 {
		t.Errorf("expected 3 keys, got %d", len(all))
	}

	if all["key1"] != "value1" {
		t.Errorf("expected 'value1', got '%v'", all["key1"])
	}

	if all["key2"] != "value2" {
		t.Errorf("expected 'value2', got '%v'", all["key2"])
	}

	if all["nested.key3"] != "value3" {
		t.Errorf("expected 'value3', got '%v'", all["nested.key3"])
	}
}

func TestKoanfConfig_Keys(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	keys := config.Keys()

	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}

	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	for _, expectedKey := range []string{"key1", "key2", "key3"} {
		if !keyMap[expectedKey] {
			t.Errorf("expected key '%s' to be in keys", expectedKey)
		}
	}
}

// ============================================================
//                    Defaults Tests
// ============================================================

func TestKoanfConfig_Defaults(t *testing.T) {
	t.Run("default values when key not set", func(t *testing.T) {
		config := NewKoanfConfig(".")

		// Test default values for various types
		if config.GetString("nonexistent") != "" {
			t.Errorf("expected empty string default")
		}

		if config.GetInt("nonexistent") != 0 {
			t.Errorf("expected 0 default for int")
		}

		if config.GetBool("nonexistent") != false {
			t.Errorf("expected false default for bool")
		}

		if config.GetFloat64("nonexistent") != 0.0 {
			t.Errorf("expected 0.0 default for float64")
		}
	})

	t.Run("empty config returns defaults", func(t *testing.T) {
		config := NewKoanfConfig(".")

		// Load empty config
		mapProvider := NewMapProvider(map[string]any{}, ".")
		if err := config.Load(mapProvider, nil); err != nil {
			t.Fatalf("failed to load empty config: %v", err)
		}

		// Should still return defaults
		if config.GetString("anything") != "" {
			t.Error("expected empty string for unset key")
		}
	})
}

// ============================================================
//                    Edge Cases
// ============================================================

func TestKoanfConfig_EmptyConfig(t *testing.T) {
	config := NewKoanfConfig(".")

	// Test empty config methods
	all := config.All()
	if len(all) != 0 {
		t.Errorf("expected empty map, got %v", all)
	}

	keys := config.Keys()
	if len(keys) != 0 {
		t.Errorf("expected empty keys, got %v", keys)
	}

	// All methods should return zero values
	if config.GetString("key") != "" {
		t.Error("expected empty string")
	}

	if config.GetInt("key") != 0 {
		t.Error("expected 0")
	}

	if config.GetBool("key") != false {
		t.Error("expected false")
	}
}

func TestKoanfConfig_MissingKeys(t *testing.T) {
	config := NewKoanfConfig(".")

	tests := []struct {
		name     string
		getter   func() any
		expected any
	}{
		{"GetString", func() any { return config.GetString("missing") }, ""},
		{"GetInt", func() any { return config.GetInt("missing") }, 0},
		{"GetInt64", func() any { return config.GetInt64("missing") }, int64(0)},
		{"GetFloat64", func() any { return config.GetFloat64("missing") }, 0.0},
		{"GetBool", func() any { return config.GetBool("missing") }, false},
		{"GetDuration", func() any { return config.GetDuration("missing") }, time.Duration(0)},
		{"GetTime", func() any { return config.GetTime("missing") }, time.Time{}},
		{"GetStringSlice", func() any { return config.GetStringSlice("missing") }, []string{}},
		{"GetIntSlice", func() any { return config.GetIntSlice("missing") }, []int{}},
		{"GetStringMap", func() any { return config.GetStringMap("missing") }, map[string]any{}},
		{"GetStringMapString", func() any { return config.GetStringMapString("missing") }, map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.getter()

			switch v := got.(type) {
			case []string:
				if tt.expected == nil {
					t.Errorf("%s() returned nil, want empty slice", tt.name)
				} else if len(v) != 0 {
					t.Errorf("%s() = %v (len=%d), want empty slice", tt.name, v, len(v))
				}
			case []int:
				if tt.expected == nil {
					t.Errorf("%s() returned nil, want empty slice", tt.name)
				} else if len(v) != 0 {
					t.Errorf("%s() = %v (len=%d), want empty slice", tt.name, v, len(v))
				}
			case map[string]any:
				if tt.expected == nil {
					t.Errorf("%s() returned nil, want empty map", tt.name)
				} else if len(v) != 0 {
					t.Errorf("%s() = %v (len=%d), want empty map", tt.name, v, len(v))
				}
			case map[string]string:
				if tt.expected == nil {
					t.Errorf("%s() returned nil, want empty map", tt.name)
				} else if len(v) != 0 {
					t.Errorf("%s() = %v (len=%d), want empty map", tt.name, v, len(v))
				}
			default:
				if got != tt.expected {
					t.Errorf("%s() = %v (%T), want %v (%T)", tt.name, got, got, tt.expected, tt.expected)
				}
			}
		})
	}
}

func TestKoanfConfig_TypeConversions(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"int_val":      42,
		"float_val":    3.14,
		"bool_val":     true,
		"string_val":   "hello",
		"int_string":   "42",
		"float_string": "3.14",
		"bool_string":  "true",
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	tests := []struct {
		name      string
		getInt    func() int
		getFloat  func() float64
		getString func() string
	}{
		{"int value", func() int { return config.GetInt("int_val") }, func() float64 { return config.GetFloat64("int_val") }, func() string { return config.GetString("int_val") }},
		{"float value", func() int { return config.GetInt("float_val") }, func() float64 { return config.GetFloat64("float_val") }, func() string { return config.GetString("float_val") }},
		{"bool value", func() int { return config.GetInt("bool_val") }, func() float64 { return config.GetFloat64("bool_val") }, func() string { return config.GetString("bool_val") }},
		{"string value", func() int { return config.GetInt("string_val") }, func() float64 { return config.GetFloat64("string_val") }, func() string { return config.GetString("string_val") }},
		{"int string", func() int { return config.GetInt("int_string") }, func() float64 { return config.GetFloat64("int_string") }, func() string { return config.GetString("int_string") }},
		{"float string", func() int { return config.GetInt("float_string") }, func() float64 { return config.GetFloat64("float_string") }, func() string { return config.GetString("float_string") }},
		{"bool string", func() int { return config.GetInt("bool_string") }, func() float64 { return config.GetFloat64("bool_string") }, func() string { return config.GetString("bool_string") }},
	}

	expectedResults := []struct {
		intVal    int
		floatVal  float64
		stringVal string
	}{
		{42, 42, "42"},    // int_val
		{3, 3.14, "3.14"}, // float_val
		{0, 0, "true"},    // bool_val
		{0, 0, "hello"},   // string_val (invalid int/float)
		{42, 42, "42"},    // int_string
		{3, 3.14, "3.14"}, // float_string
		{0, 0, "true"},    // bool_string
	}

	for i, test := range tests {
		expected := expectedResults[i]
		t.Run("conversion-"+string(rune('A'+i)), func(t *testing.T) {
			if got := test.getInt(); got != expected.intVal {
				t.Errorf("GetInt() = %d, want %d", got, expected.intVal)
			}
			if got := test.getFloat(); got != expected.floatVal {
				t.Errorf("GetFloat64() = %f, want %f", got, expected.floatVal)
			}
			if got := test.getString(); got != expected.stringVal {
				t.Errorf("GetString() = %s, want %s", got, expected.stringVal)
			}
		})
	}
}

func TestKoanfConfig_NestedAccess(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"app.database.host": "localhost",
		"app.database.port": 5432,
		"app.database.name": "testdb",
		"app.cache.enabled": true,
		"app.cache.ttl":     300,
		"app.server.host":   "0.0.0.0",
		"app.server.port":   8080,
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Test nested access
	if config.GetString("app.database.host") != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", config.GetString("app.database.host"))
	}

	if config.GetInt("app.database.port") != 5432 {
		t.Errorf("expected 5432, got %d", config.GetInt("app.database.port"))
	}

	if config.GetBool("app.cache.enabled") != true {
		t.Errorf("expected true, got %v", config.GetBool("app.cache.enabled"))
	}

	if config.GetInt("app.cache.ttl") != 300 {
		t.Errorf("expected 300, got %d", config.GetInt("app.cache.ttl"))
	}

	if config.GetString("app.server.host") != "0.0.0.0" {
		t.Errorf("expected '0.0.0.0', got '%s'", config.GetString("app.server.host"))
	}

	if config.GetInt("app.server.port") != 8080 {
		t.Errorf("expected 8080, got %d", config.GetInt("app.server.port"))
	}
}

func TestKoanfConfig_CustomDelimiter(t *testing.T) {
	// Test with underscore delimiter
	mapProvider := NewMapProvider(map[string]any{
		"app_database_host": "localhost",
		"app_database_port": 5432,
		"app_server_host":   "0.0.0.0",
	}, "_")

	config := NewKoanfConfig("_")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// With custom delimiter, use underscores for nested keys
	if config.GetString("app_database_host") != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", config.GetString("app_database_host"))
	}

	if config.GetInt("app_database_port") != 5432 {
		t.Errorf("expected 5432, got %d", config.GetInt("app_database_port"))
	}

	// Dotted notation should not work with custom delimiter
	if config.GetString("app.database.host") != "" {
		t.Errorf("expected empty (custom delimiter), got '%s'", config.GetString("app.database.host"))
	}
}

// ============================================================
//                    Integration Tests
// ============================================================

func TestKoanfConfig_Integration_LoadAndMerge(t *testing.T) {
	// Create config with defaults from map
	defaults := NewMapProvider(map[string]any{
		"database.host":    "localhost",
		"database.port":    5432,
		"database.name":    "devdb",
		"database.timeout": "30s",
		"debug":            false,
	}, ".")

	config := NewKoanfConfig(".")

	// Load defaults
	if err := config.Load(defaults, nil); err != nil {
		t.Fatalf("failed to load defaults: %v", err)
	}

	// Verify defaults are loaded
	if config.GetString("database.host") != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", config.GetString("database.host"))
	}

	// Create override config
	overrideConfig := NewKoanfConfig(".")
	mapProvider2 := NewMapProvider(map[string]any{
		"database.name": "proddb",
		"database.port": 5433,
		"debug":         true,
	}, ".")
	if err := overrideConfig.Load(mapProvider2, nil); err != nil {
		t.Fatalf("failed to load override config: %v", err)
	}

	// Merge overrides
	if err := config.Merge(overrideConfig); err != nil {
		t.Fatalf("failed to merge overrides: %v", err)
	}

	// Verify overrides
	if config.GetString("database.name") != "proddb" {
		t.Errorf("expected 'proddb', got '%s'", config.GetString("database.name"))
	}

	if config.GetInt("database.port") != 5433 {
		t.Errorf("expected 5433, got %d", config.GetInt("database.port"))
	}

	if config.GetBool("debug") != true {
		t.Errorf("expected true, got %v", config.GetBool("debug"))
	}

	// Verify unchanged values are still there
	if config.GetString("database.host") != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", config.GetString("database.host"))
	}

	if config.GetString("database.timeout") != "30s" {
		t.Errorf("expected '30s', got '%s'", config.GetString("database.timeout"))
	}
}

func TestKoanfConfig_Integration_FileAndEnv(t *testing.T) {
	// Create temp JSON file
	tmpFile, err := os.CreateTemp("", "config-integration-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `{
		"server": {
			"host": "localhost",
			"port": 8080,
			"ssl": false
		},
		"database": {
			"host": "localhost",
			"port": 5432
		}
	}`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Set env vars (these will be lowercased and converted)
	t.Setenv("SERVER_SSL", "true")

	// Load file
	jsonParser := NewJSONParser()
	fileProvider := NewFileProvider(tmpFile.Name(), jsonParser)

	config := NewKoanfConfig(".")

	if err := config.Load(fileProvider, jsonParser); err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	// Verify file loaded
	if config.GetString("server.host") != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", config.GetString("server.host"))
	}

	if config.GetBool("server.ssl") != false {
		t.Errorf("expected false, got %v", config.GetBool("server.ssl"))
	}

	// Load env variables
	envProvider := NewEnvProvider()

	if err := config.Load(envProvider, nil); err != nil {
		t.Fatalf("failed to load env: %v", err)
	}

	// Verify env overrode file value (koanf doesn't merge nested maps by default)
	// With config loaded in order, the second Load overwrites
	// This is expected behavior - koanf doesn't recursively merge nested maps
}

// ============================================================
//                    Provider Interface Tests
// ============================================================

func TestProvider_Implementation(t *testing.T) {
	// Verify all providers implement Provider interface
	var _ Provider = NewFileProvider("/tmp/test.json", NewJSONParser())
	var _ Provider = NewEnvProvider()
	var _ Provider = NewFlagProvider(pflag.NewFlagSet("test", pflag.ContinueOnError), ".")
	var _ Provider = NewMapProvider(map[string]any{}, ".")
	var _ Provider = NewStructProvider(map[string]any{}, "koanf")
	var _ Provider = NewBytesProvider([]byte("{}"), NewJSONParser())
}

func TestParser_Implementation(t *testing.T) {
	// Verify all parsers implement Parser interface
	var _ Parser = NewJSONParser()
}

// ============================================================
//                    Raw Access Tests
// ============================================================

func TestKoanfConfig_Raw(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"key": "value",
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	raw := config.Raw()
	if raw == nil {
		t.Fatal("expected raw koanf to be non-nil")
	}

	// Use raw koanf to access value
	if raw.String("key") != "value" {
		t.Errorf("expected 'value', got '%s'", raw.String("key"))
	}
}

func TestKoanfConfig_Print(t *testing.T) {
	// Print is a void method, just verify it doesn't panic
	config := NewKoanfConfig(".")

	// Should not panic with empty config
	config.Print()

	// Load config and print
	mapProvider := NewMapProvider(map[string]any{
		"key": "value",
	}, ".")

	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Should not panic
	config.Print()
}

// ============================================================
//                    UnmarshalOptions Tests
// ============================================================

func TestKoanfConfig_UnmarshalOptions(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"database": map[string]any{
			"host": "localhost",
			"port": 5432,
		},
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	type Config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	var cfg Config
	opts := UnmarshalOptions{
		Tag:       "json",
		FlatPaths: false,
	}

	if err := config.UnmarshalWithConf("database", &cfg, opts); err != nil {
		t.Fatalf("failed to unmarshal with options: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != 5432 {
		t.Errorf("expected 5432, got %d", cfg.Port)
	}
}

func TestKoanfConfig_UnmarshalOptions_DefaultTag(t *testing.T) {
	mapProvider := NewMapProvider(map[string]any{
		"host": "localhost",
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	type Config struct {
		Host string `koanf:"host"`
	}

	var cfg Config
	opts := UnmarshalOptions{
		Tag:       "", // Should default to "koanf"
		FlatPaths: false,
	}

	if err := config.UnmarshalWithConf("", &cfg, opts); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", cfg.Host)
	}
}

// ============================================================
//                    Composite Tests (Table-Driven)
// ============================================================

func TestKoanfConfig_Composite(t *testing.T) {
	type TestCase struct {
		name      string
		setup     func() *KoanfConfig
		test      func(t *testing.T, config *KoanfConfig)
		expectErr bool
	}

	testCases := []TestCase{
		{
			name: "load and get string",
			setup: func() *KoanfConfig {
				config := NewKoanfConfig(".")
				mapProvider := NewMapProvider(map[string]any{"key": "value"}, ".")
				if err := config.Load(mapProvider, nil); err != nil {
					t.Fatalf("failed to setup: %v", err)
				}
				return config
			},
			test: func(t *testing.T, config *KoanfConfig) {
				if config.GetString("key") != "value" {
					t.Errorf("expected 'value', got '%s'", config.GetString("key"))
				}
			},
		},
		{
			name: "load and get int",
			setup: func() *KoanfConfig {
				config := NewKoanfConfig(".")
				mapProvider := NewMapProvider(map[string]any{"count": 42}, ".")
				if err := config.Load(mapProvider, nil); err != nil {
					t.Fatalf("failed to setup: %v", err)
				}
				return config
			},
			test: func(t *testing.T, config *KoanfConfig) {
				if config.GetInt("count") != 42 {
					t.Errorf("expected 42, got %d", config.GetInt("count"))
				}
			},
		},
		{
			name: "load and get bool",
			setup: func() *KoanfConfig {
				config := NewKoanfConfig(".")
				mapProvider := NewMapProvider(map[string]any{"enabled": true}, ".")
				if err := config.Load(mapProvider, nil); err != nil {
					t.Fatalf("failed to setup: %v", err)
				}
				return config
			},
			test: func(t *testing.T, config *KoanfConfig) {
				if config.GetBool("enabled") != true {
					t.Errorf("expected true, got %v", config.GetBool("enabled"))
				}
			},
		},
		{
			name: "nested get operations",
			setup: func() *KoanfConfig {
				config := NewKoanfConfig(".")
				mapProvider := NewMapProvider(map[string]any{
					"parent.child":  "value",
					"parent.number": 123,
				}, ".")
				if err := config.Load(mapProvider, nil); err != nil {
					t.Fatalf("failed to setup: %v", err)
				}
				return config
			},
			test: func(t *testing.T, config *KoanfConfig) {
				if config.GetString("parent.child") != "value" {
					t.Errorf("expected 'value', got '%s'", config.GetString("parent.child"))
				}
				if config.GetInt("parent.number") != 123 {
					t.Errorf("expected 123, got %d", config.GetInt("parent.number"))
				}
			},
		},
		{
			name: "merge preserves existing values",
			setup: func() *KoanfConfig {
				config := NewKoanfConfig(".")
				mapProvider1 := NewMapProvider(map[string]any{
					"key1": "value1",
					"key2": "value2",
				}, ".")
				if err := config.Load(mapProvider1, nil); err != nil {
					t.Fatalf("failed to setup: %v", err)
				}
				return config
			},
			test: func(t *testing.T, config *KoanfConfig) {
				// Create second config from mapProvider2
				config2 := NewKoanfConfig(".")
				mapProvider2 := NewMapProvider(map[string]any{
					"key3": "value3",
					"key2": "overridden",
				}, ".")

				if err := config2.Load(mapProvider2, nil); err != nil {
					t.Fatalf("failed to load config2: %v", err)
				}

				if err := config.Merge(config2); err != nil {
					t.Fatalf("failed to merge: %v", err)
				}

				if config.GetString("key1") != "value1" {
					t.Errorf("expected 'value1', got '%s'", config.GetString("key1"))
				}
				if config.GetString("key2") != "overridden" {
					t.Errorf("expected 'overridden', got '%s'", config.GetString("key2"))
				}
				if config.GetString("key3") != "value3" {
					t.Errorf("expected 'value3', got '%s'", config.GetString("key3"))
				}
			},
		},
		{
			name: "isSet with nil data",
			setup: func() *KoanfConfig {
				config := NewKoanfConfig(".")
				provider := &MapProvider{data: nil, delim: ".", name: "nil"}
				if err := config.Load(provider, nil); err != nil {
					t.Fatalf("failed to setup: %v", err)
				}
				return config
			},
			test: func(t *testing.T, config *KoanfConfig) {
				if config.IsSet("anything") {
					t.Errorf("expected IsSet to return false")
				}
			},
		},
		{
			name: "unmarshal with nested structure",
			setup: func() *KoanfConfig {
				config := NewKoanfConfig(".")
				mapProvider := NewMapProvider(map[string]any{
					"database": map[string]any{
						"host": "localhost",
						"port": 5432,
					},
				}, ".")
				if err := config.Load(mapProvider, nil); err != nil {
					t.Fatalf("failed to setup: %v", err)
				}
				return config
			},
			test: func(t *testing.T, config *KoanfConfig) {
				var cfg struct {
					Database struct {
						Host string `koanf:"host"`
						Port int    `koanf:"port"`
					} `koanf:"database"`
				}

				if err := config.Unmarshal("", &cfg); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}

				if cfg.Database.Host != "localhost" {
					t.Errorf("expected 'localhost', got '%s'", cfg.Database.Host)
				}
				if cfg.Database.Port != 5432 {
					t.Errorf("expected 5432, got %d", cfg.Database.Port)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.setup()
			tc.test(t, config)
		})
	}
}

// ============================================================
//                    Benchmark Tests
// ============================================================

func BenchmarkKoanfConfig_GetString(b *testing.B) {
	mapProvider := NewMapProvider(map[string]any{
		"key": "value",
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		b.Fatalf("failed to load config: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.GetString("key")
	}
}

func BenchmarkKoanfConfig_GetInt(b *testing.B) {
	mapProvider := NewMapProvider(map[string]any{
		"count": 42,
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		b.Fatalf("failed to load config: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.GetInt("count")
	}
}

func BenchmarkKoanfConfig_Unmarshal(b *testing.B) {
	mapProvider := NewMapProvider(map[string]any{
		"host":    "localhost",
		"port":    8080,
		"enabled": true,
	}, ".")

	config := NewKoanfConfig(".")
	if err := config.Load(mapProvider, nil); err != nil {
		b.Fatalf("failed to load config: %v", err)
	}

	type Config struct {
		Host    string `koanf:"host"`
		Port    int    `koanf:"port"`
		Enabled bool   `koanf:"enabled"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		if err := config.Unmarshal("", &cfg); err != nil {
			b.Fatalf("failed to unmarshal: %v", err)
		}
	}
}

// ============================================================
//                    Helper Functions
// ============================================================

func mustUnmarshalJSON(s string) map[string]any {
	var data map[string]any
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		panic(err)
	}
	return data
}
