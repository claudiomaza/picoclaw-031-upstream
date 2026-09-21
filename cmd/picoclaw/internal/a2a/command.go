package a2a

import (
	"fmt"
	"log"
	"net/http"
	"os"

	ext "github.com/sipeed/picoclaw/pkg/extensibility/a2a"
	"github.com/spf13/cobra"
)

func NewA2ACommand() *cobra.Command {
	var configPath, addr string
	cmd := &cobra.Command{Use: "a2a", Short: "Run the PicoClaw A2A runtime profile"}
	serve := &cobra.Command{
		Use:   "serve",
		Short: "Serve the A2A profile over HTTP",
		RunE: func(*cobra.Command, []string) error {
			if configPath == "" {
				configPath = os.Getenv("PICOCLAW_CONFIG")
			}
			if configPath == "" {
				configPath = os.ExpandEnv("$HOME/.picoclaw/config.json")
			}
			runner, err := ext.NewRunner(configPath)
			if err != nil {
				return err
			}
			if addr == "" {
				addr = os.Getenv("PICOCLAW_PROFILE_HTTP_ADDR")
			}
			if addr == "" {
				addr = "127.0.0.1:8644"
			}
			log.Printf("PicoClaw A2A profile listening on %s", addr)
			return http.ListenAndServe(addr, runner.ProfileHandler())
		},
	}
	serve.Flags().StringVar(&configPath, "config", "", "PicoClaw configuration path")
	serve.Flags().StringVar(&addr, "addr", "", "HTTP listen address")
	cmd.AddCommand(serve)
	cmd.AddCommand(&cobra.Command{Use: "version", Run: func(*cobra.Command, []string) { fmt.Println("PicoClaw A2A profile") }})
	return cmd
}
