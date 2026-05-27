package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	awsAdapter "monostack/internal/adapters/aws"
	"monostack/internal/adapters/tui"
	configPkg "monostack/internal/pkg/config"
	"monostack/internal/usecase"
)

var (
	version = "0.0.1"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("monostack %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built at: %s\n", date)
		return
	}

	cfgStore := configPkg.NewFileConfigStore("config.json")
	subStore := configPkg.NewFileSubscriptionStore("subscriptions.json")
	cfgUseCase := usecase.NewConfigUseCaseWithSubscriptions(cfgStore, subStore)

	s3Adapter := awsAdapter.NewS3Adapter()
	sqsAdapter := awsAdapter.NewSQSAdapter()
	snsAdapter := awsAdapter.NewSNSAdapter()
	secretsAdapter := awsAdapter.NewSecretsAdapter()

	awsUseCase := usecase.NewAWSUseCase(s3Adapter, sqsAdapter, snsAdapter, secretsAdapter)
	snapshotUseCase := usecase.NewSnapshotUseCase(awsUseCase, cfgUseCase)

	tui.Version = version
	m := tui.NewModel(awsUseCase, cfgUseCase, snapshotUseCase)
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "monostack: %v\n", err)
		os.Exit(1)
	}
}
