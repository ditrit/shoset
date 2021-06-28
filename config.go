package shoset

import "github.com/spf13/viper"

type Config struct {
    ShosetJoin []string `mapstructure:"SHOSETJOIN"`
	ShosetLink []string `mapstructure:"SHOSETLINK"`
}

func LoadConfig(path string) (config Config, err error) {
    viper.AddConfigPath(path)
    viper.SetConfigName("config_shoset")
    viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
    if err != nil {
        return
    }

    err = viper.Unmarshal(&config)
    return
}