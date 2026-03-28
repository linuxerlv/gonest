package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// FileProvider 文件配置提供者
type FileProvider struct {
	path   string
	name   string
	parser Parser
}

// NewFileProvider 创建文件提供者
func NewFileProvider(path string, parser Parser) *FileProvider {
	return &FileProvider{
		path:   path,
		name:   fmt.Sprintf("file:%s", path),
		parser: parser,
	}
}

func (p *FileProvider) Read() (map[string]any, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", p.path, err)
	}
	return p.parser.Parse(data)
}

func (p *FileProvider) Watch(callback func(any, error)) error {
	f := file.Provider(p.path)
	return f.Watch(func(event any, err error) {
		callback(event, err)
	})
}

func (p *FileProvider) Name() string {
	return p.name
}

// EnvProvider 环境变量配置提供者
type EnvProvider struct {
	prefix      string
	delimiter   string
	transformFn func(string, string) (string, any)
	name        string
}

// EnvOption 环境变量选项
type EnvOption func(*EnvProvider)

// WithEnvPrefix 设置环境变量前缀
func WithEnvPrefix(prefix string) EnvOption {
	return func(p *EnvProvider) {
		p.prefix = prefix
	}
}

// WithEnvDelimiter 设置分隔符
func WithEnvDelimiter(delimiter string) EnvOption {
	return func(p *EnvProvider) {
		p.delimiter = delimiter
	}
}

// WithEnvTransform 设置转换函数
func WithEnvTransform(fn func(string, string) (string, any)) EnvOption {
	return func(p *EnvProvider) {
		p.transformFn = fn
	}
}

// NewEnvProvider 创建环境变量提供者
func NewEnvProvider(opts ...EnvOption) *EnvProvider {
	p := &EnvProvider{
		prefix:    "",
		delimiter: ".",
		name:      "environment",
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *EnvProvider) Read() (map[string]any, error) {
	transformFn := p.transformFn
	if transformFn == nil {
		transformFn = func(k, v string) (string, any) {
			k = strings.ToLower(strings.TrimPrefix(k, p.prefix))
			k = strings.ReplaceAll(k, "_", p.delimiter)
			return k, v
		}
	}

	provider := env.Provider(p.prefix, p.delimiter, func(s string) string {
		k, _ := transformFn(s, "")
		return k
	})

	var k = koanf.New(p.delimiter)
	if err := k.Load(provider, nil); err != nil {
		return nil, err
	}
	return k.All(), nil
}

func (p *EnvProvider) Watch(callback func(any, error)) error {
	return fmt.Errorf("env provider does not support watch")
}

func (p *EnvProvider) Name() string {
	return p.name
}

// FlagProvider 命令行参数配置提供者
type FlagProvider struct {
	flagSet *pflag.FlagSet
	delim   string
	name    string
}

// NewFlagProvider 创建命令行参数提供者
func NewFlagProvider(flagSet *pflag.FlagSet, delim string) *FlagProvider {
	return &FlagProvider{
		flagSet: flagSet,
		delim:   delim,
		name:    "flags",
	}
}

func (p *FlagProvider) Read() (map[string]any, error) {
	var k = koanf.New(p.delim)
	provider := posflag.Provider(p.flagSet, p.delim, k)
	if err := k.Load(provider, nil); err != nil {
		return nil, err
	}
	return k.All(), nil
}

func (p *FlagProvider) Watch(callback func(any, error)) error {
	return fmt.Errorf("flag provider does not support watch")
}

func (p *FlagProvider) Name() string {
	return p.name
}

// MapProvider Map配置提供者
type MapProvider struct {
	data  map[string]any
	delim string
	name  string
}

// NewMapProvider 创建Map提供者
func NewMapProvider(data map[string]any, delim string) *MapProvider {
	return &MapProvider{
		data:  data,
		delim: delim,
		name:  "map",
	}
}

func (p *MapProvider) Read() (map[string]any, error) {
	if p.data == nil {
		return map[string]any{}, nil
	}
	return p.data, nil
}

func (p *MapProvider) Watch(callback func(any, error)) error {
	return fmt.Errorf("map provider does not support watch")
}

func (p *MapProvider) Name() string {
	return p.name
}

// StructProvider 结构体配置提供者
type StructProvider struct {
	data any
	tag  string
	name string
}

// NewStructProvider 创建结构体提供者
func NewStructProvider(data any, tag string) *StructProvider {
	return &StructProvider{
		data: data,
		tag:  tag,
		name: "struct",
	}
}

func (p *StructProvider) Read() (map[string]any, error) {
	var k = koanf.New(".")
	provider := structs.Provider(p.data, p.tag)
	if err := k.Load(provider, nil); err != nil {
		return nil, err
	}
	return k.All(), nil
}

func (p *StructProvider) Watch(callback func(any, error)) error {
	return fmt.Errorf("struct provider does not support watch")
}

func (p *StructProvider) Name() string {
	return p.name
}

// BytesProvider 字节数组配置提供者
type BytesProvider struct {
	data   []byte
	parser Parser
	name   string
}

// NewBytesProvider 创建字节数组提供者
func NewBytesProvider(data []byte, parser Parser) *BytesProvider {
	return &BytesProvider{
		data:   data,
		parser: parser,
		name:   "bytes",
	}
}

func (p *BytesProvider) Read() (map[string]any, error) {
	return p.parser.Parse(p.data)
}

func (p *BytesProvider) Watch(callback func(any, error)) error {
	return fmt.Errorf("bytes provider does not support watch")
}

func (p *BytesProvider) Name() string {
	return p.name
}

// JSONParser JSON解析器
type JSONParser struct{}

func NewJSONParser() *JSONParser { return &JSONParser{} }

func (p *JSONParser) Parse(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

func (p *JSONParser) Marshal(data map[string]any) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

func (p *JSONParser) Name() string { return "json" }

// YAMLParser YAML解析器
type YAMLParser struct{}

func NewYAMLParser() *YAMLParser { return &YAMLParser{} }

func (p *YAMLParser) Parse(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return result, nil
}

func (p *YAMLParser) Marshal(data map[string]any) ([]byte, error) {
	return yaml.Marshal(data)
}

func (p *YAMLParser) Name() string { return "yaml" }

// KoanfFileProvider koanf文件提供者包装
func KoanfFileProvider(path string) Provider {
	return &koanfFileProvider{path: path}
}

type koanfFileProvider struct {
	path string
}

func (p *koanfFileProvider) Read() (map[string]any, error) {
	var k = koanf.New(".")
	if err := k.Load(file.Provider(p.path), nil); err != nil {
		return nil, err
	}
	return k.All(), nil
}

func (p *koanfFileProvider) Watch(callback func(any, error)) error {
	return fmt.Errorf("use file.Provider directly for watch support")
}

func (p *koanfFileProvider) Name() string {
	return fmt.Sprintf("koanf-file:%s", p.path)
}

// KoanfConfMapProvider koanf配置Map提供者
func KoanfConfMapProvider(data map[string]any, delim string) Provider {
	return &koanfConfMapProvider{data: data, delim: delim}
}

type koanfConfMapProvider struct {
	data  map[string]any
	delim string
}

func (p *koanfConfMapProvider) Read() (map[string]any, error) {
	var k = koanf.New(p.delim)
	if err := k.Load(confmap.Provider(p.data, p.delim), nil); err != nil {
		return nil, err
	}
	return k.All(), nil
}

func (p *koanfConfMapProvider) Watch(callback func(any, error)) error {
	return fmt.Errorf("confmap provider does not support watch")
}

func (p *koanfConfMapProvider) Name() string {
	return "koanf-confmap"
}

// KoanfRawBytesProvider koanf原始字节提供者
func KoanfRawBytesProvider(data []byte) Provider {
	return &koanfRawBytesProvider{data: data}
}

type koanfRawBytesProvider struct {
	data []byte
}

func (p *koanfRawBytesProvider) Read() (map[string]any, error) {
	return nil, nil // 需要配合Parser使用
}

func (p *koanfRawBytesProvider) Watch(callback func(any, error)) error {
	return fmt.Errorf("rawbytes provider does not support watch")
}

func (p *koanfRawBytesProvider) Name() string {
	return "koanf-rawbytes"
}

// 内部使用
func rawBytesProvider(data []byte) *rawbytes.RawBytes {
	return rawbytes.Provider(data)
}
