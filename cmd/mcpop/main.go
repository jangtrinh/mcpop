package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/jangtrinh/mcpop/internal/analyzer"
	"github.com/jangtrinh/mcpop/internal/proxy"
	"github.com/jangtrinh/mcpop/internal/server"
	"github.com/jangtrinh/mcpop/internal/storage"
)

const version = "0.1.0"

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	root := &cobra.Command{
		Use:           "mcpop",
		Short:         "Observability proxy and silent-failure catcher for MCP servers",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newRunCmd(), newDashboardCmd(), newVersionCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRunCmd() *cobra.Command {
	var (
		port          int
		bind          string
		slowThreshold int
		noUI          bool
		debug         bool
	)

	cmd := &cobra.Command{
		Use:   "run -- <command> [args...]",
		Short: "Wrap an MCP server, log JSON-RPC traffic, and surface failures",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProxy(runOptions{
				command:       args[0],
				args:          args[1:],
				port:          port,
				bind:          bind,
				slowThreshold: slowThreshold,
				noUI:          noUI,
				debug:         debug,
			})
		},
	}

	cmd.Flags().IntVar(&port, "port", 4040, "dashboard port")
	cmd.Flags().StringVar(&bind, "bind", "127.0.0.1", "dashboard bind address")
	cmd.Flags().IntVar(&slowThreshold, "slow-threshold", 5000, "slow tool threshold in milliseconds")
	cmd.Flags().BoolVar(&noUI, "no-ui", false, "log to SQLite only, do not start the dashboard")
	cmd.Flags().BoolVar(&debug, "debug", false, "log dropped events and non-JSON-RPC lines to stderr")
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func newDashboardCmd() *cobra.Command {
	var (
		port int
		bind string
	)

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Open the dashboard against previously recorded sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := storage.Open("")
			if err != nil {
				return err
			}
			defer db.Close()

			repo := storage.NewRepository(db)
			hub := server.NewSSEHub()
			return serveDashboard(repo, hub, bind, port, true)
		},
	}

	cmd.Flags().IntVar(&port, "port", 4040, "dashboard port")
	cmd.Flags().StringVar(&bind, "bind", "127.0.0.1", "dashboard bind address")
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}

type runOptions struct {
	command       string
	args          []string
	port          int
	bind          string
	slowThreshold int
	noUI          bool
	debug         bool
}

func runProxy(opts runOptions) error {
	db, err := storage.Open("")
	if err != nil {
		return err
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	ctx := context.Background()

	argv := append([]string{opts.command}, opts.args...)
	session := &storage.Session{Command: strings.Join(argv, " ")}
	if err := repo.CreateSession(ctx, session); err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	hub := server.NewSSEHub()
	engine := analyzer.NewEngine(repo, int64(opts.slowThreshold))
	engine.SetOnFailureCallback(func(f *storage.Failure) {
		hub.Broadcast("failure", f)
	})

	p := proxy.NewStdioProxy(repo, session.ID, opts.command, opts.args, opts.debug)
	p.SetAnalyzer(engine)
	p.SetSSEHub(hub)

	if !opts.noUI {
		srv := server.NewServer(repo, hub, opts.port)
		srv.SetBind(opts.bind)
		go func() {
			if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("[mcpop] dashboard failed: %v", err)
			}
		}()
		defer func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = srv.Stop(stopCtx)
		}()
		log.Printf("[mcpop] dashboard http://%s:%d", srv.Bind(), srv.Port())
	}

	return p.Run(ctx)
}

func serveDashboard(repo *storage.Repository, hub *server.SSEHub, bind string, port int, block bool) error {
	srv := server.NewServer(repo, hub, port)
	srv.SetBind(bind)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	log.Printf("[mcpop] dashboard http://%s:%d", srv.Bind(), srv.Port())
	if !block {
		return nil
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-sig:
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Stop(stopCtx)
	}
}
