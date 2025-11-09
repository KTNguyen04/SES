package main

import (
	"flag"
	"fmt"
	"log"

	ses "github.com/KTNguyen04/SES/internal/protocol"
	"github.com/spf13/viper"
)

var (
	port = flag.Int("port", 9000, "Self port")
)

func main() {
	flag.Parse()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var cfg struct {
		Processes []struct {
			Id   int    `mapstructure:"id"`
			Host string `mapstructure:"host"`
			Port int    `mapstructure:"port"`
		} `mapstructure:"processes"`
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}

	host := ses.NewHost(-1, "127.0.0.1", *port)

	peers := &ses.Peers{}
	for _, p := range cfg.Processes {
		if p.Port == *port {
			host.Id = p.Id
			continue
		}
		peers.AddPeer(p.Id, p.Host, p.Port)
	}
	fmt.Printf("%v", peers.Peers)
	fmt.Printf("Self: %v\n", host)

}
