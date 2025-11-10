package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/KTNguyen04/SES/internal/p2p"
	"github.com/spf13/viper"
)

var (
	port = flag.String("port", "9000", "Self port")
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
			Port string `mapstructure:"port"`
		} `mapstructure:"processes"`
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}

	host := p2p.NewHost(-1, "127.0.0.1", *port)
	knownPeers := []p2p.Peer{}

	for _, p := range cfg.Processes {
		if p.Port == *port {
			host.Id = p.Id
			continue
		}
		knownPeers = append(knownPeers, p2p.Peer{
			Id:      p.Id,
			Address: p.Host,
			Port:    p.Port,
		})
	}
	fmt.Printf("%v", knownPeers)
	fmt.Printf("Self: %v\n", host)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		host.RunServer()
	}()

	for _, peer := range knownPeers {
		host.DialToPeer(peer.Address, peer.Port)
		host.Inform(peer.Address, peer.Port)
	}
	defer host.ClosePeerConnection()
	wg.Wait()

}
