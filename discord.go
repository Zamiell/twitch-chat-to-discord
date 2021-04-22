package main

import (
	"os"

	"github.com/bwmarrin/discordgo"
)

var (
	discord                *discordgo.Session
	discordGuildID         string
	discordOutputChannelID string
	discordEmojiMap        map[string]string
	discordIsConnected     = false
)

func discordInit() {
	// Read some configuration values from environment variables
	discordToken := os.Getenv("DISCORD_TOKEN")
	if len(discordToken) == 0 {
		logger.Fatal("The \"DISCORD_TOKEN\" environment variable is blank.")
		return
	}
	discordGuildID = os.Getenv("DISCORD_GUILD_ID")
	if len(discordGuildID) == 0 {
		logger.Fatal("The \"DISCORD_GUILD_ID\" environment variable is blank.")
		return
	}
	discordOutputChannelID = os.Getenv("DISCORD_OUTPUT_CHANNEL_ID")
	if len(discordOutputChannelID) == 0 {
		logger.Fatal("The \"DISCORD_OUTPUT_CHANNEL_ID\" environment variable is blank.")
		return
	}

	// Bot accounts must be prefixed with "Bot"
	if v, err := discordgo.New("Bot " + discordToken); err != nil {
		logger.Errorf("Failed to create a Discord session: %s", err.Error())
		return
	} else {
		discord = v
	}

	// Register function handlers for various events
	discord.AddHandler(discordReady)

	// Open the websocket and begin listening
	if err := discord.Open(); err != nil {
		logger.Fatalf("Failed to open the Discord session: %s", err.Error())
		return
	}
}

func discordReady(s *discordgo.Session, event *discordgo.Ready) {
	logger.Infof("Discord bot connected with username: %s", event.User.Username)
	discordGetEmojiIDs(s)
	discordIsConnected = true
}

func discordGetEmojiIDs(s *discordgo.Session) {
	// Get the emoji IDs for the various Twitch badges
	var guildEmojis []*discordgo.Emoji
	if v, err := s.GuildEmojis(discordGuildID); err != nil {
		logger.Fatalf("Failed to get the emojis for the Discord guild: %s", err.Error())
		return
	} else {
		guildEmojis = v
	}

	discordEmojiMap = make(map[string]string)
	for _, badgeName := range twitchBadges {
		foundMatchingEmoji := false
		for _, emoji := range guildEmojis {
			if emoji.Name == badgeName {
				foundMatchingEmoji = true
				discordEmojiMap[badgeName] = emoji.ID
				break
			}
		}
		if !foundMatchingEmoji {
			logger.Fatalf(
				"Failed to find the matching Discord emoji for the Twitch badge of: %s",
				badgeName,
			)
		}
	}
}

func discordSend(to string, msg string) {
	if !discordIsConnected {
		return
	}

	if _, err := discord.ChannelMessageSend(to, msg); err != nil {
		// Occasionally, sending messages to Discord can time out; if this occurs,
		// do not bother retrying, since losing a single message is fairly meaningless
		logger.Infof("Failed to send \"%s\" to Discord: %s", msg, err.Error())
		return
	}
}
