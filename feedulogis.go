package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type FilterFunction func(*discordgo.Message) bool

const defaultPrefix = "!fdl"
const FilterNone = 0
const FilterGab = 2
const FilterGabCommands = 4

var token string

func init() {

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()

	if token == "" {
		fmt.Println("No token")
		os.Exit(1)
	}
}

func main() {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	// Cleanly close down the Discord session on return.
	defer dg.Close()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Feedulogis is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	s.UpdateStatus(0, defaultPrefix+"help")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by other bots or the bot itself
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	// check if the message starts with defined prefix
	if !strings.HasPrefix(m.Content, defaultPrefix) {
		return
	}

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		fmt.Println("Pas de channel: ", err)
		return
	}
	aperm, err := s.UserChannelPermissions(m.Author.ID, c.ID)
	if err != nil {
		// Could not find permissions.
		fmt.Println("Pas de permissions: ", err)
		return
	}

	if aperm&discordgo.PermissionManageMessages == 0 {
		s.ChannelMessageSend(c.ID, "Pas assez de permissions pour ça.")
		return
	}

	args := strings.Split(m.Content, " ")
	prefixCommand := args[0]
	args = args[1:]
	fmt.Printf("%s : %s\n", m.Author.Username, prefixCommand)
	commandName := strings.Replace(prefixCommand, defaultPrefix, "", 1)
	message := strings.Join(args, " ")

	fmt.Println(commandName, message)

	switch commandName {
	case "clean", "c", "javel", "j":
		var offset = 1
		var scanIndex = 0
		var beforeID = ""
		if len(args) == 3 && args[0] == "before" {
			beforeID = args[1]
			scanIndex = 2
			offset = 0
		} else {

		}
		nbToScan, err := strconv.Atoi(args[scanIndex])
		if err != nil {
			s.ChannelMessageSend(c.ID, "Syntaxe invalide")
			fmt.Println("Syntaxe invalide: ", err)
			return
		}



		// il faut prendre en compte le message de suppression, on ajoute 1 a nbToScan
		messages, err := s.ChannelMessages(c.ID, nbToScan+offset, beforeID, "", "")
		if err != nil {
			s.ChannelMessageSend(c.ID, "Erreur lors de la recuperation des messages")
			fmt.Println("Erreur lors de la recuperation des messages: ", err)
			return
		}
		fmt.Println(len(messages))
		var filterFunction FilterFunction
		switch commandName {
		case "clean", "c":
			filterFunction = filterGabCommands
		case  "javel", "j":
			filterFunction = nil

		}
		messageIds := getMessagesIdsToDelete(messages, filterFunction)

		err = s.ChannelMessagesBulkDelete(c.ID, messageIds)
		if err != nil {
			s.ChannelMessageSend(c.ID, "Erreur lors de la suppression des messages")
			fmt.Println("Erreur lors de la suppression des messages: ", err)
			return
		}
		switch commandName {
		case "clean", "c":
			s.ChannelMessageSend(c.ID, "Et voilà! Tout Propre! J'ai trié " +
				strconv.Itoa(nbToScan) + " messages")

		case "javel", "j":
			s.ChannelMessageSend(c.ID, "Et voilà! Tout Propre! J'ai supprimé " +
				strconv.Itoa(nbToScan) + " messages")
		}

	case "help":
		s.ChannelMessageSend(c.ID,
			defaultPrefix+"<clean|javel> [before <id du message>] <nb de message à suppr(max 100)>")
	}
}

func getMessagesIdsToDelete(messages []*discordgo.Message, filterFunc FilterFunction) (messageIds []string) {
	for _, message := range messages {
		if filterFunc != nil && filterFunc(message) {
			messageIds = append(messageIds, message.ID)
		}
	}
	return messageIds
}

func filterGabCommands(message *discordgo.Message) bool {
	if message.Author.ID != "415147492745936897" &&
		!strings.HasPrefix(message.Content, "!gab") {
		return true
	}
	return false
}

func filterGab(message *discordgo.Message) bool {
	if message.Author.ID != "415147492745936897" {
		return true
	}
	return false
}