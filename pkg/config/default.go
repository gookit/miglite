package config


var std *Config

// Get returns the default configuration instance
func Get() *Config {
	if std == nil {
		panic("default config not load and initialized")
	}
	return std
}

// SetDefault the default configuration instance
func SetDefault(c *Config) { std = c }

// Reset the default configuration instance
func Reset() { std = nil }

// LoadDefault load configuration to the default instance
func LoadDefault(configPath string) error {
	var err error
	std, err = Load(configPath)
	if err != nil {
		return err
	}
	return nil
}
