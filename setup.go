package setup

import (
	"log"

	"github.com/spf13/viper"
)

// Config contains and provides the configuration that is required at runtime
type Config interface {
	GetString(string) string
	GetInt(string) int
	GetInt64(string) int64
	GetBool(string) bool
}

// GetConfig sets up the configuration by environment variables and default values
func GetConfig() (Config, error) {
	// load a configuration from the same folder where this method is called
	viper.AddConfigPath(".")
	viper.SetConfigName("core")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	// ----- Env bindings -----

	// Router env
	viper.BindEnv("http.base.address", "HTTP_BASE_ADDRESS")
	viper.SetDefault("http.base.address", "0.0.0.0")

	viper.BindEnv("http.port", "HTTP_PORT")
	viper.SetDefault("http.port", 8080)

	if err := viper.ReadInConfig(); err != nil {
		if err, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("mc-template-generator core config file not found, falling back to default and ENV")
		} else {
			// Config file was found but another error was produced
			return nil, err
		}
	} else {
		log.Println("mc-template-generator core config file found")
	}

	return viper.GetViper(), nil
}
