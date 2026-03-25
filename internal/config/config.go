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
	"github.com/spf13/viper"
	"golang.org/x/mod/module"
)

var dirRegex = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)

type Config struct {
	Version      string          `mapstructure:"-"`
	Filename     string          `mapstructure:"-"`
	Log          *Log            `mapstructure:"log"`
	Uncountables []string        `mapstructure:"uncountables"`
	Initialisms  []string        `mapstructure:"initialisms"`
	Timezone     string          `mapstructure:"timezone"`
	Template     *Template       `mapstructure:"template"`
	Output       *Output         `mapstructure:"output"`
	Objects      []*CustomObject `mapstructure:"objects"`
	DB           []*DB           `mapstructure:"db"`
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
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

type Template struct {
	Dir   string `mapstructure:"dir"`
	Model string `mapstructure:"model"`
}

func (t *Template) Validate() error {
	if t.Model == "" {
		return errors.New("template.model is required")
	}

	return nil
}

type Output struct {
	Dir     string `mapstructure:"dir"`
	Package string `mapstructure:"package"`
	Module  string `mapstructure:"module"`
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
	From string `mapstructure:"from"`
	To   string `mapstructure:"to"`
}

type CustomObject struct {
	Name    string         `mapstructure:"name"`
	Comment string         `mapstructure:"comment"`
	Fields  []*CustomField `mapstructure:"fields"`
	Tags    []*Tag         `mapstructure:"tags"`
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
	Import   string `mapstructure:"import"`
	Name     string `mapstructure:"name"`
	Type     string `mapstructure:"type"`
	Relation bool   `mapstructure:"relation"` // true for relation
	Tags     []*Tag `mapstructure:"tags"`
	Comment  string `mapstructure:"comment"`
}

func (c *CustomField) Validate() error {
	if c.Name == "" {
		return errors.New("custom.fields.name is required")
	}
	if err := validateIdent(c.Name); err != nil {
		return fmt.Errorf("custom.fields.name %w", err)
	}
	if c.Type == "" {
		return errors.New("custom.fields.type is required")
	}
	if err := validateIdent(c.Type); err != nil {
		return fmt.Errorf("custom.fields.type %w", err)
	}
	return nil
}

type DB struct {
	Driver   string                  `mapstructure:"driver"`
	DSN      string                  `mapstructure:"dsn"`
	Schema   string                  `mapstructure:"schema"`
	Included []string                `mapstructure:"included"`
	Excluded []string                `mapstructure:"excluded"`
	Customs  map[string]*CustomTable `mapstructure:"customs"`
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
	Name    string                  `mapstructure:"table"`
	Alias   string                  `mapstructure:"alias"`
	Comment string                  `mapstructure:"comment"`
	Fields  map[string]*CustomField `mapstructure:"fields"`
	Tags    []*Tag                  `mapstructure:"tags"`
}

func (c *CustomTable) Validate() error {
	if c.Name == "" {
		return errors.New("custom.table.name is required")
	}
	if len(c.Fields) == 0 {
		return errors.New("custom.table.fields is required")
	}
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
		config.DB = []*DB{
			{
				DSN: args.DSN,
			},
		}
	} else {
		if args.DSN != "" {
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
	v := viper.New()
	v.SetConfigFile(actualFileName)
	v.AddConfigPath(".")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}
	config.Filename = actualFileName
	inflection.AddUncountable(config.Uncountables...)
	return &config, nil
}

func loadENV() (string, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return v.GetString("env"), nil
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
