package config

type Tag struct {
	Key     string   `mapstructure:"key"`
	Name    string   `mapstructure:"name"`
	Options []string `mapstructure:"options"`
}
