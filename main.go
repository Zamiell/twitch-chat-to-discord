package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	ProjectName = "twitch-chat-to-discord"
)

var (
	logger      *zap.SugaredLogger
	projectPath string
)

func main() {
	// Initialize logging using the Zap library
	var zapLogger *zap.Logger
	if v, err := zap.NewDevelopment(); err != nil {
		log.Fatalf("Failed to initialize logging: %s", err.Error())
	} else {
		zapLogger = v
	}
	logger = zapLogger.Sugar()

	// Welcome message
	startText := "| Starting " + ProjectName + " |"
	borderText := "+" + strings.Repeat("-", len(startText)-2) + "+" // nolint: gomnd
	logger.Info(borderText)
	logger.Info(startText)
	logger.Info(borderText)

	// Get the project path
	// https://stackoverflow.com/questions/18537257/
	if v, err := os.Executable(); err != nil {
		logger.Fatal("Failed to get the path of the currently running executable: %s", err.Error())
	} else {
		projectPath = filepath.Dir(v)
	}

	// Check to see if the ".env" file exists
	envPath := path.Join(projectPath, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		logger.Fatalf(
			"The \"%s\" file does not exist. Copy the \".env_template\" file to \".env\".",
			envPath,
		)
		return
	} else if err != nil {
		logger.Fatal("Failed to check if the \"%s\" file exists: %s", envPath, err.Error())
		return
	}

	// Load the ".env" file which contains environment variables with secret values
	if err := godotenv.Load(envPath); err != nil {
		logger.Fatal("Failed to load the \".env\" file: %s", err.Error())
		return
	}

	// Initialize the Discord connection (in "discord.go")
	discordInit()
	defer discord.Close()

	// Initialize the Twitch connection (in "twitch.go")
	go twitchInit()

	// Block until a terminal signal is received
	logger.Infof("%s is now running.", ProjectName)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Flush the logging buffer (we don't care if this fails)
	defer logger.Sync()
}
