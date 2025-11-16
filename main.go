package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	comm "github.com/KTNguyen04/SES/communication"
	"github.com/KTNguyen04/SES/internal/p2p"
	"github.com/spf13/viper"
)

var (
	port = flag.String("port", "9000", "Self port")
)

func main() {
	flag.Parse()

	timestamp := time.Now().Format("2006-01-02_15:04:05")
	logDir := fmt.Sprintf("./logs/%s", *port)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}
	logFileName := fmt.Sprintf("./logs/%v/%s-chat_server.log", *port, timestamp)

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

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
	host.Vvt.V = make([]*comm.Vector, len(cfg.Processes))
	fmt.Printf("vvT length: %d\n", len(host.Vvt.V))

	knownPeers := []p2p.Peer{}

	for i, p := range cfg.Processes {
		if p.Port == *port {
			host.Id = p.Id
			host.Vvt.V[i] = &comm.Vector{
				T: make([]int64, len(cfg.Processes)),
			}

			continue
		}
		host.Vvt.V[i] = &comm.Vector{
			T: nil,
		}
		knownPeers = append(knownPeers, p2p.Peer{
			Id:      p.Id,
			Address: p.Host,
			Port:    p.Port,
		})

	}

	var wg sync.WaitGroup
	wg.Go(
		func() {
			host.RunServer()
		},
	)

	for _, peer := range knownPeers {
		p := peer

		// Connect to all peers
		wg.Go(func() {
			host.DialToPeer(p)
			host.Pinging(p)
		})
	}

	// Wait for all connections to be established
	log.Printf("Waiting for all connections to be established...")
	time.Sleep(10 * time.Second)
	log.Printf("Start sending messages...")

	speed := viper.GetInt("no_of_messages_per_minute")
	interval := time.Duration(60/speed) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	total := viper.GetInt("no_of_messages")

	// Send messages to all peers
	for _, peer := range knownPeers {
		p := peer
		wg.Go(func() {
			for i := 1; i <= total; i++ {
				<-ticker.C
				host.SESSendMessage(p.Id, fmt.Sprintf("Message %d from %d to %d", i, host.Id, p.Id))
			}
		})
	}

	wg.Wait()

}
