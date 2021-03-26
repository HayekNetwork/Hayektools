package config

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/koding/multiconfig"
)

// 配置文件
type Config struct {
	DebugMode bool `default:"false"` // 调试模式(打印内部日志)

	Host        string `default:"http://127.0.0.1:8585"` // 主链地址(RPC服务)
	UserName    string `default:"Hayek"`
	UserKey     string `default:""`
	UserAddress string `default:"0x3eb41fc94f240242c9bbb8bf46b9feb356fd09e2"`

	XUserAddressBook map[string]string // 其它地址簿 map[name]address
}

func Default() *Config {
	conf := new(Config)

	loader := newWithPath("")
	if err := loader.Load(conf); err != nil {
		panic(err)
	}

	return conf
}

func Load(path string) (*Config, error) {
	conf := new(Config)

	loader := newWithPath(path)
	if err := loader.Load(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func MustLoad(path string) *Config {
	conf := new(Config)

	loader := newWithPath(path)
	if err := loader.Load(conf); err != nil {
		panic(err)
	}

	return conf
}

func (m *Config) GetAddress(id string) string {
	if len(m.XUserAddressBook) > 0 {
		if v, ok := m.XUserAddressBook[id]; ok {
			return v
		}
	}
	return id
}

func (m *Config) Clone() *Config {
	var q = *m
	return &q
}

func (m *Config) JSONString() string {
	if b, err := json.MarshalIndent(m, "", "\t"); err == nil {
		return string(b) + "\n"
	}
	return ""
}

func (m *Config) TOMLString() string {
	buf := new(bytes.Buffer)
	buf.WriteString("# HayekTool\n\n")
	if err := toml.NewEncoder(buf).Encode(m); err != nil {
		panic(err)
	}
	s := buf.String()
	s = strings.ReplaceAll(s, "\n", "\r\n")
	return s
}

func newWithPath(path string) *multiconfig.DefaultLoader {
	var loaders []multiconfig.Loader

	// Read default values defined via tag fields "default"
	loaders = append(loaders, &multiconfig.TagLoader{})

	// Choose what while is passed
	if strings.HasSuffix(path, ".toml") {
		loaders = append(loaders, &multiconfig.TOMLLoader{Path: path})
	}

	if strings.HasSuffix(path, ".json") {
		loaders = append(loaders, &multiconfig.JSONLoader{Path: path})
	}

	loader := multiconfig.MultiLoader(loaders...)

	d := &multiconfig.DefaultLoader{}
	d.Loader = loader
	d.Validator = multiconfig.MultiValidator(&multiconfig.RequiredValidator{})
	return d
}
