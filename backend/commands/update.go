package commands

import (
	"fmt"
	"os"
	"vendel/services"

	"github.com/spf13/cobra"
)

// NewUpdateCommand returns the `update` CLI subcommand.
func NewUpdateCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Check for updates and self-update the binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("  Current version:  %s\n", version)
			fmt.Printf("  Platform:         %s\n\n", services.DetectPlatform())

			if version == "dev" || version == "docker" || version == "" {
				return fmt.Errorf("self-update is not available in %s mode", version)
			}

			if isRunningInDocker() {
				return fmt.Errorf("self-update is not supported inside Docker; use `docker compose pull && docker compose up -d` instead")
			}
			if isRunningInNixOS() {
				return fmt.Errorf("self-update is not supported on NixOS; manage vendel through your nix configuration")
			}

			fmt.Print("[1/3] Checking for updates... ")
			services.InvalidateCache()
			latest, err := services.CheckLatest(version)
			if err != nil {
				fmt.Println("FAILED")
				return fmt.Errorf("failed to check: %w", err)
			}

			if latest.Version == version {
				fmt.Println("up to date.")
				return nil
			}
			fmt.Printf("v%s available\n", latest.Version)

			if latest.ReleaseURL != "" {
				fmt.Printf("       Release notes: %s\n", latest.ReleaseURL)
			}

			fmt.Print("[2/3] Downloading and verifying... ")
			binaryPath, err := services.DownloadAndVerify(latest.AssetURL, latest.Checksum)
			if err != nil {
				fmt.Println("FAILED")
				return fmt.Errorf("download failed: %w", err)
			}
			fmt.Println("done")

			fmt.Print("[3/3] Replacing binary... ")
			if err := services.ApplyUpdate(binaryPath); err != nil {
				fmt.Println("FAILED")
				return fmt.Errorf("apply failed: %w", err)
			}
			fmt.Println("done")

			fmt.Printf("\n  Updated: v%s → v%s\n\n", version, latest.Version)
			fmt.Println("  Restart the service to apply the update.")
			return nil
		},
	}
}

func isRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func isRunningInNixOS() bool {
	_, err := os.Stat("/etc/NIXOS")
	return err == nil
}
