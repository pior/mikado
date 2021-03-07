# Mikado

**This is a work in progress**

- Automatic dependency injection (reflection based)
- Service management based on the [pior/runnable] interface
- Service lifecycle depends on dependencies
  - Start after the dependencies, stop before them
- Concurrent service start and stop

Limitations:
- Variadic parameters are not supported in constructors

### Example:

```go
package main

import (
	"context"

	"github.com/pior/runnable"
)

type Config struct {
	DatabaseHost string
}

func BuildConfig(cliOption *CLIOption) *Config {
	return &Config{DatabaseHost: "127.0.0.1"}
}

type Database struct{}

func NewDatabase(cfg *Config) *Database {
	return &Database{}
}

func (d *Database) Run(ctx context.Context) error {
	// database has started
	<-ctx.Done()
	// database has stopped
	return nil
}

type Server struct{}

func NewServer(cfg *Config, cli *CLIOption, db *Database) *Server {
	return &Server{}
}

func NewServerBase() *Server {
	return &Server{}
}

func (s *Server) Run(ctx context.Context) error {
	// server is doing something
	<-ctx.Done()
	// server has stopped
	return nil
}

func Main() {
	a := mikado.New()

	a.AddProvider(BuildConfig)
	a.AddRunnable(NewDatabase)
	a.AddRunnable(NewServer)

    runnable.Run(a) // err := a.Run(context.Background())
}
```
