package bot

import (
	"github.com/bwmarrin/discordgo"
)

func HasUser(users []*discordgo.User, target *discordgo.User) bool {
	for _, user := range users {
		if user.ID == target.ID {
			return true
		}
	}
	return false
}
