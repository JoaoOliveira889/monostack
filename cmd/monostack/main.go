package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	awsAdapter "monostack/internal/adapters/aws"
	"monostack/internal/adapters/tui"
	configPkg "monostack/internal/pkg/config"
	"monostack/internal/usecase"
)

var (
	version = "0.0.9"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	configPath := flag.String("config", "", "Path to configuration file (default: ~/.config/monostack/config.json)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("monostack %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built at: %s\n", date)
		return
	}

	var cfgStore *configPkg.FileConfigStore
	if *configPath != "" {
		cfgStore = configPkg.NewFileConfigStoreFromPath(*configPath)
	} else {
		cfgStore = configPkg.NewFileConfigStore("config.json")
	}
	subStore := configPkg.NewFileSubscriptionStore("subscriptions.json")
	cfgUseCase := usecase.NewConfigUseCaseWithSubscriptions(cfgStore, subStore)

	cache := awsAdapter.NewClientCache()
	s3Adapter := awsAdapter.NewS3Adapter(cache)
	sqsAdapter := awsAdapter.NewSQSAdapter(cache)
	snsAdapter := awsAdapter.NewSNSAdapter(cache)
	secretsAdapter := awsAdapter.NewSecretsAdapter(cache)

	awsUseCase := usecase.NewAWSUseCase(s3Adapter, sqsAdapter, snsAdapter, secretsAdapter)
	snapshotUseCase := usecase.NewSnapshotUseCase(awsUseCase, cfgUseCase)

	tui.Version = version
	m := tui.NewModel(awsUseCase, cfgUseCase, snapshotUseCase)
	p := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	m.SetProgram(p)

	origStderr := os.Stderr
	sdkReader, sdkWriter, _ := os.Pipe()
	os.Stderr = sdkWriter
	log.SetOutput(sdkWriter)

	go func() {
		scanner := bufio.NewScanner(sdkReader)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				select {
				case m.LogCh() <- line:
				default:
				}
			}
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(origStderr, "monostack: %v\n", err)
		os.Exit(1)
	}

	sdkWriter.Close()
	sdkReader.Close()
}
