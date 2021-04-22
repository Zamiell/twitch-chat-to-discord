package main

import (
	"fmt"
	"os"

	"github.com/gempir/go-twitch-irc/v2"
)

var (
	twitchUsername     string
	twitchInputChannel string
	twitchClient       *twitch.Client

	// https://help.twitch.tv/s/article/twitch-chat-badges-guide?language=en_US
	twitchBadges = []string{
		"staff",
		"admin",
		"moderator",
		"verified",
		"vip",
	}
)

func twitchInit() {
	// Read some configuration values from environment variables
	twitchUsername = os.Getenv("TWITCH_USERNAME")
	if len(twitchUsername) == 0 {
		logger.Fatal("The \"TWITCH_USERNAME\" environment variable is blank.")
		return
	}
	twitchOAuth := os.Getenv("TWITCH_OAUTH")
	if len(twitchOAuth) == 0 {
		logger.Fatal("The \"TWITCH_OAUTH\" environment variable is blank.")
		return
	}
	twitchInputChannel = os.Getenv("TWITCH_INPUT_CHANNEL")
	if len(twitchOAuth) == 0 {
		logger.Fatal("The \"TWITCH_INPUT_CHANNEL\" environment variable is blank.")
		return
	}

	twitchClient = twitch.NewClient(twitchUsername, twitchOAuth)

	twitchClient.OnConnect(twitchReady)
	twitchClient.OnPrivateMessage(twitchMessage)

	// The "Connect()" method is blocking
	if err := twitchClient.Connect(); err != nil {
		logger.Errorf("Failed to create a Twitch session: %s", err.Error())
	}
}

func twitchReady() {
	logger.Infof("Twitch bot connected with username: %s", twitchUsername)
	logger.Infof("Joining Twitch channel: %s", twitchInputChannel)
	twitchClient.Join(twitchInputChannel)
}

func twitchMessage(message twitch.PrivateMessage) {
	logger.Infof("<%s> %s", message.User.Name, message.Message)

	// Emulate Twitch badge images with Discord emoji
	// The badge value is the version, which we don't need
	// See: https://dev.twitch.tv/docs/irc/tags
	badgeEmoji := ""

	for _, badgeName := range twitchBadges {
		if _, ok := message.User.Badges[badgeName]; !ok {
			continue
		}

		var emojiID string
		if v, ok := discordEmojiMap[badgeName]; !ok {
			logger.Errorf("Failed to find the emoji ID for the Twitch badge of: %s", badgeName)
			continue
		} else {
			emojiID = v
		}

		badgeEmoji += fmt.Sprintf("<:%s:%s>", badgeName, emojiID)
	}

	discordMessage := fmt.Sprintf("%s <**%s**> %s", badgeEmoji, message.User.Name, message.Message)
	discordSend(discordOutputChannelID, discordMessage)
}
