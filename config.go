package main

import (
    "github.com/spf13/viper"
    "fmt"
    "sync"
    "errors"
)

// type Config struct {
//     ShosetJoin string `mapstructure:"SHOSETJOIN"`
// 	ShosetLink string `mapstructure:"SHOSETLINK"`
//     m sync.Mutex
// }

// func LoadConfig(path string) (config Config, err error) {
//     viper.AddConfigPath(path)
//     viper.SetConfigName("config_shoset")
//     viper.SetConfigType("yaml")
// 	viper.AutomaticEnv()

// 	err = viper.ReadInConfig()
//     if err != nil {
//         return
//     }

//     err = viper.Unmarshal(&config)
//     return
// }

// // func SetConfig(config *Config) {
// //     config.m.Lock()
// // 	defer config.m.Unlock()
    
// //     table := []string{"hallo", "hello"}
// //     table = append(table, "salut")
// //     viper.Set(config.ShosetJoin, table)
// //     viper.WriteConfig()
// // }

// func SetConfig(config *Config, mode, value string) error {
//     config.m.Lock()
// 	defer config.m.Unlock()
    
//     if mode == "join" {
//         viper.Set(config.ShosetJoin, value)
//     } else if mode == "link" {
//         viper.Set(config.ShosetLink, value)
//     } else {
//         fmt.Println("error : wrong input mode in SetConfig")
//         return errors.New("error : wrong input mode in SetConfig")
//     }
    
//     viper.WriteConfig()
//     return nil
// }

// func main() {
//     config, err := LoadConfig(".")
//     if err != nil {
//         fmt.Println("Error in viper config : ", err)
//         return
//     }
//     SetConfig(&config, "join", "+++")
//     SetConfig(&config, "lnk", "---")
//     fmt.Println(viper.Get(config.ShosetJoin))
//     fmt.Println(viper.Get(config.ShosetLink)) 
// }

type Config struct {
    ShosetJoin string `mapstructure:"SHOSETJOIN"`
	ShosetLink string `mapstructure:"SHOSETLINK"`
    m sync.Mutex
}

func LoadConfig(path string) (config Config, err error) {
    viper.AddConfigPath(path)
    viper.SetConfigName("config_shoset")
    viper.SetConfigType("yaml")
	viper.AutomaticEnv()

    viper.SetDefault("ShosetJoin", "")
    viper.SetDefault("ShosetLink", "")

	err = viper.ReadInConfig()
    if err != nil {
        return
    }

    err = viper.Unmarshal(&config)
    return
}

func SetConfig(config *Config, mode, value string) error {
    config.m.Lock()
	defer config.m.Unlock()
    
    if mode == "join" {
        viper.Set(config.ShosetJoin, value)
    } else if mode == "link" {
        viper.Set(config.ShosetLink, value)
    } else {
        fmt.Println("error : wrong input mode in SetConfig")
        return errors.New("error : wrong input mode in SetConfig")
    }
    
    viper.WriteConfig()
    return nil
}

func GetConfig(config *Config) map[string][]string {
    config.m.Lock()
	defer config.m.Unlock()

    m := make(map[string][]string)
    m["ShosetJoin"] = viper.GetStringSlice("ShosetJoin")
    m["ShosetLink"] = viper.GetStringSlice("ShosetLink")
    return m
}

func main() {
    config, err := LoadConfig(".")
    if err != nil {
        fmt.Println("Error in viper config : ", err)
        return
    }
    
    SetConfig(&config, "join", "+++")
    SetConfig(&config, "link", "---")
    my_map := GetConfig(&config)
    fmt.Println(my_map)
}