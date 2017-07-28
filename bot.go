package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"strings"
	"syscall"
)


func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string

func main() {
	if token == "" {
		fmt.Println("No token provided. Please run: snob-bot -t <bot token>")
		return
	}
	dg, err := discordgo.New("Bot " + token)
	dg.AddHandler(messageCreate)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if strings.HasPrefix(m.Content, "!help") {
		s.ChannelMessageSend(m.ChannelID,"Receiving help will only weaken you")
	}
	if strings.HasPrefix(m.Content, "!join") {
		err := listenChannel(s,m)

		if err != nil {
			s.ChannelMessageSend(m.ChannelID,fmt.Sprintf("Cannot join channel: %v",err))
		}
	}

}
func listenChannel(s *discordgo.Session, m *discordgo.MessageCreate) error {
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return nil
	}
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		// Could not find guild.
		return nil
	}
	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			_, err := s.ChannelVoiceJoin(g.ID, vs.ChannelID, true, true)
			return err
		}
	}
	return nil
}
