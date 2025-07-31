package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/GregVetkin/logtracker/internal/archiver"
	"github.com/GregVetkin/logtracker/internal/tracker"
	"github.com/GregVetkin/logtracker/pkg/logger"
)

func main() {
	logger := logger.NewLogger(os.Stdout, logger.LevelInfo)

	config := tracker.Config{
		FilesToWatch: []string{
			"C:/Users/Drydad/Desktop/test1.txt", 
			"C:/Users/Drydad/Desktop/test2.txt"},
	}

	fileTracker := tracker.New(config, logger)
	archiver := archiver.NewZipArchiver("logs_archive.zip", logger)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if err := fileTracker.Start(); err != nil {
		logger.Fatal("Failed to start tracker", "error", err)
	}

	<-sigChan
	logger.Info("Shutting down...")

	fileTracker.Stop()
	if err := archiver.Archive(fileTracker.NewFiles()); err != nil {
		logger.Error("Archiving failed", "error", err)
	}
}