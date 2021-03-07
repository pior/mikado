package mikado

import (
	"context"
	"flag"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type CLIOption struct {
	configFile string
}

func ParseCLI() (*CLIOption, error) {
	configFile := flag.String("config", "~/.appconfig", "path to the config file")
	flag.Parse()

	return &CLIOption{*configFile}, nil
}

type Config struct {
	DatabaseHost string
}

func BuildConfig(cliOption *CLIOption) *Config {
	// use cliOption.configFile to load the config
	return &Config{
		DatabaseHost: "127.0.0.1",
	}
}

type Store interface {
	List() []string
}

type MemoryStore struct{}

func NewMemoryStore() Store {
	return &MemoryStore{}
}

func (d *MemoryStore) List() []string {
	return []string{"one", "two"}
}

func (d *MemoryStore) Run(ctx context.Context) error {
	log.Print("database has started")
	<-ctx.Done()
	log.Print("database has stopped")
	return nil
}

type Server struct{}

func NewServer(cfg *Config, cli *CLIOption, store Store) *Server {
	return &Server{}
}

func (s *Server) Run(ctx context.Context) error {
	log.Print("server is doing something")
	<-ctx.Done()
	log.Print("server has stopped")
	return nil
}

func Test_App(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	a := New()
	a.AddProvider(ParseCLI)
	a.AddProvider(BuildConfig)
	a.AddRunnable(NewMemoryStore)
	a.AddRunnable(NewServer)

	err := a.Run(ctx)
	require.NoError(t, err)
}
