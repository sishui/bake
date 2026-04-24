// Package config provides configuration for the application.
package config

import (
	"errors"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"golang.org/x/mod/module"
)

var dirRegex = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)

type Config struct {
	Version      string          `koanf:"-"`
	Filename     string          `koanf:"-"`
	Log          *Log            `koanf:"log"`
	Uncountables []string        `koanf:"uncountables"`
	Initialisms  []string        `koanf:"initialisms"`
	Timezone     string          `koanf:"timezone"`
	Template     *Template       `koanf:"template"`
	Output       *Output         `koanf:"output"`
	Objects      []*CustomObject `koanf:"objects"`
	DB           []*DB           `koanf:"db"`
}

func (c *Config) Validate() error {
	if err := c.Template.Validate(); err != nil {
		return err
	}

	if err := c.Output.Validate(); err != nil {
		return err
	}

	for _, db := range c.DB {
		if err := db.Validate(); err != nil {
			return err
		}
	}

	for _, o := range c.Objects {
		if err := o.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) Naming() map[string]string {
	result := make(map[string]string, len(c.Initialisms))
	for _, s := range c.Initialisms {
		result[strings.ToLower(s)] = s
	}

	return result
}

type Log struct {
	Level string `koanf:"level"`
	File  string `koanf:"file"`
}

type Template struct {
	Dir   string `koanf:"dir"`
	Model string `koanf:"model"`
}

func (t *Template) Validate() error {
	if t.Model == "" {
		return errors.New("template.model is required")
	}

	return nil
}

type Output struct {
	Dir     string `koanf:"dir"`
	Package string `koanf:"package"`
	Module  string `koanf:"module"`
}

func (o *Output) Validate() error {
	if o.Dir == "" {
		return errors.New("output.dir is required")
	}

	if !dirRegex.MatchString(o.Dir) {
		return errors.New("output.dir is invalid")
	}

	if err := validatePackage(o.Package); err != nil {
		return fmt.Errorf("output.package %w", err)
	}
	if err := module.CheckPath(o.Module); err != nil {
		return fmt.Errorf("output.module %w", err)
	}

	return nil
}

type Replacement struct {
	From string `koanf:"from"`
	To   string `koanf:"to"`
}

type CustomObject struct {
	Name    string         `koanf:"name"`
	Comment string         `koanf:"comment"`
	Fields  []*CustomField `koanf:"fields"`
	Tags    []*Tag         `koanf:"tags"`
}

func (c *CustomObject) Validate() error {
	if c.Name == "" {
		return errors.New("custom.name is required")
	}

	if len(c.Fields) == 0 {
		return errors.New("custom.fields is required")
	}

	for _, field := range c.Fields {
		if err := field.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type CustomField struct {
	Import   string `koanf:"import"`
	Name     string `koanf:"name"`
	Type     string `koanf:"type"`
	Relation bool   `koanf:"relation"` // true for relation
	Tags     []*Tag `koanf:"tags"`
	Comment  string `koanf:"comment"`
}

func (c *CustomField) Validate() error {
	if c.Name != "" {
		if err := validateIdent(c.Name); err != nil {
			return fmt.Errorf("custom.fields.name %w", err)
		}
	}

	return nil
}

type DB struct {
	Driver   string                  `koanf:"driver"`
	DSN      string                  `koanf:"dsn"`
	Schema   string                  `koanf:"schema"`
	Included []string                `koanf:"included"`
	Excluded []string                `koanf:"excluded"`
	Customs  map[string]*CustomTable `koanf:"customs"`
}

func (d *DB) Validate() error {
	if d.Driver == "" {
		return errors.New("db.driver is required")
	}
	if d.DSN == "" {
		return errors.New("db.dsn is required")
	}
	switch d.Driver {
	case "mysql":
	case "postgres":
		if d.Schema == "" {
			return errors.New("db.schema is required")
		}
	default:
		return errors.New("db.driver must be mysql or postgres")
	}
	for _, custom := range d.Customs {
		if err := custom.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type CustomTable struct {
	Name    string                  `koanf:"table"`
	Alias   string                  `koanf:"alias"`
	Comment string                  `koanf:"comment"`
	Fields  map[string]*CustomField `koanf:"fields"`
	Tags    []*Tag                  `koanf:"tags"`
}

func (c *CustomTable) Validate() error {
	for _, field := range c.Fields {
		if err := field.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *CustomTable) CloseFields() map[string]*CustomField {
	if c == nil {
		return nil
	}
	fields := make(map[string]*CustomField, len(c.Fields))
	for _, field := range c.Fields {
		fields[field.Name] = field
	}
	return fields
}

func Parse(args Args) (*Config, error) {
	config, err := parse(args.Config)
	if err != nil {
		return nil, err
	}
	if config.Output == nil {
		config.Output = &Output{}
	}
	if args.Output != "" {
		config.Output.Dir = args.Output
	}
	if args.Package != "" {
		config.Output.Package = args.Package
	} else if config.Output.Package == "" {
		config.Output.Package = filepath.Base(config.Output.Dir)
	}
	if config.DB == nil {
		if args.DSN == "" {
			return nil, errors.New("db.dsn is required: set via config file or --dsn flag")
		}
		driver := detectDriver(args.DSN)
		config.DB = []*DB{
			{
				Driver: driver,
				DSN:    args.DSN,
			},
		}
	} else {
		if args.DSN != "" {
			config.DB[0].Driver = detectDriver(args.DSN)
			config.DB[0].DSN = args.DSN
		}
	}
	if config.Template == nil {
		config.Template = &Template{
			Model: "model",
		}
	}
	if config.Log == nil {
		config.Log = &Log{
			Level: "info",
		}
	}
	if args.Verbose {
		config.Log.Level = "debug"
	}
	return config, nil
}

func parse(filename string) (*Config, error) {
	env, err := loadENV()
	if err != nil {
		return nil, err
	}
	actualFileName := filename
	index := strings.LastIndex(filename, ".gen")
	if env != "" && index != -1 {
		actualFileName = fmt.Sprintf("%s.%s%s", filename[:index], env, filename[index:])
	}
	if _, err := os.Stat(actualFileName); os.IsNotExist(err) {
		actualFileName = filename
	}
	k := koanf.New(".")
	if err := k.Load(file.Provider(actualFileName), yaml.Parser()); err != nil {
		return nil, err
	}
	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		return nil, err
	}
	config.Filename = actualFileName
	inflection.AddUncountable(config.Uncountables...)
	return &config, nil
}

func loadENV() (string, error) {
	buffer, err := os.ReadFile(".env")
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	for line := range strings.SplitSeq(string(buffer), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "env="); ok {
			return after, nil
		}
	}

	return "", nil
}

func validatePackage(name string) error {
	if err := validateIdent(name); err != nil {
		return err
	}

	if name != strings.ToLower(name) {
		return fmt.Errorf("package name must be lowercase: %s", name)
	}

	if strings.Contains(name, "_") {
		return fmt.Errorf("package name should not contain underscore: %s", name)
	}

	return nil
}

func validateIdent(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if !token.IsIdentifier(name) {
		return fmt.Errorf("invalid identifier: %s", name)
	}
	if token.Lookup(name).IsKeyword() {
		return fmt.Errorf("identifier cannot be keyword: %s", name)
	}
	return nil
}

func detectDriver(dsn string) string {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return "postgres"
	}
	return "mysql"
}
