package discord // "github.com/itszuvalex/mcdiscord/pkg/discord"
import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/itszuvalex/mcdiscord/pkg/api"
)

func (discord *DiscordHandler) handleListCommands(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	var commandfields []*discordgo.MessageEmbedField
	for _, command := range discord.Commands {
		commandfields = append(commandfields, &discordgo.MessageEmbedField{
			Name:   command.Name,
			Value:  fmt.Sprintf("Usage: %s\n%s", discord.config.ControlChar+command.Usage, command.Description),
			Inline: true,
		})
	}
	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00,
		Description: "Bot discord commands",
		Fields:      commandfields,
		Timestamp:   time.Now().Format(time.RFC3339),
		Title:       "List Commands",
	}
	_, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embed)
	if err != nil {
		fmt.Println("Error embedding message", err)
		return err
	}
	return nil
}

func (discord *DiscordHandler) handleAddServer(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	if discord.config.ChannelId != m.Message.ChannelID {
		return errors.New("Wrong channel")
	}
	args := strings.Split(data, " ")
	if len(args) < 2 {
		fmt.Println("Add server needs server name.")
		return errors.New("Add server needs args {ip:port} {name}")
	}

	location, err := api.ParseNetLocation(args[0])
	if err != nil {
		fmt.Println("Add server could not parse NetLocation:", err)
		return err
	}

	name := strings.Join(args[1:], " ")
	err = discord.serverhandler.AddServer(*location, name)
	if err != nil {
		fmt.Println("Add server could not add server", err)
		return err
	}
	return nil
}

func (discord *DiscordHandler) handleRemoveServer(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	if discord.config.ChannelId != m.Message.ChannelID {
		return errors.New("Wrong channel")
	}
	if strings.Contains(data, ":") {
		location, err := api.ParseNetLocation(data)
		if err != nil {
			return err
		}
		return discord.serverhandler.RemoveServer(*location)
	} else {
		return discord.serverhandler.RemoveServerByName(data)
	}
}

func (discord *DiscordHandler) handleSetChannel(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	discord.config.ChannelId = m.ChannelID
	return discord.masterconfig.Write()
}

func (discord *DiscordHandler) handleJsonTest(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	return nil
}

func (discord *DiscordHandler) handleListServers(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	if discord.config.ChannelId != m.Message.ChannelID {
		fmt.Println("Wrong channel")
		return errors.New("Wrong channel")
	}
	var serverfields []*discordgo.MessageEmbedField
	for _, server := range discord.serverhandler.Servers() {
		serverfields = append(serverfields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s", server.Name()),
			Value: fmt.Sprintf("%s:%d", server.Location().Address, server.Location().Port),
		})
	}
	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00,
		Description: "Servers connected to discord.",
		Fields:      serverfields,
		Timestamp:   time.Now().Format(time.RFC3339),
		Title:       "List Servers",
	}
	_, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embed)
	if err != nil {
		fmt.Println("Error embedding message", err)
		return err
	}
	return nil
}

func (discord *DiscordHandler) handleListServerCommands(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	var commandfields []*discordgo.MessageEmbedField
	for _, command := range discord.ServerCommands {
		commandfields = append(commandfields, &discordgo.MessageEmbedField{
			Name:   command.Name,
			Value:  fmt.Sprintf("Usage: %s\n%s", discord.config.ControlChar+command.Usage, command.Description),
			Inline: true,
		})
	}
	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00,
		Description: "Known server commands",
		Fields:      commandfields,
		Timestamp:   time.Now().Format(time.RFC3339),
		Title:       "List Server Commands",
	}

	_, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embed)
	if err != nil {
		fmt.Println("Error embedding message", err)
		return err
	}
	return nil
}

func (discord *DiscordHandler) handleListPermissions(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	filterstring := strings.TrimSpace(data)
	fmt.Println("Filterstring:", filterstring)

	var lines []string

	for _, perm := range discord.permhandler.RecursiveGetAllNodes() {
		if filterstring == "" || strings.HasPrefix(perm.FullName(), filterstring) {
			lines = append(lines, perm.FullName()[len(filterstring):])
		}
	}

	sort.Strings(lines)

	_, err := s.ChannelMessageSend(m.Message.ChannelID, strings.Join(lines, "\n"))
	if err != nil {
		fmt.Println("Error sending message", err)
		return err
	}
	return nil
}

func (discord *DiscordHandler) handleAddPermRole(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	ispace := strings.Index(strings.TrimSpace(data), " ")
	nodepath := data[:ispace]
	restofdata := data[ispace+1:]

	node, err := discord.permhandler.GetPermNode(nodepath)
	if err != nil {
		fmt.Println("Node not found for path:", nodepath)
		return err
	}

	roleID := ""
	rolename := ""
	// <@& %s >
	restofdata = strings.TrimSpace(restofdata)
	ibracket := strings.Index(restofdata, ">")
	if ibracket > 0 {
		rolemention := restofdata[:ibracket]
		restofdata = restofdata[ibracket+1:]
		roleID = rolemention[3:len(rolemention)]
	} else {
		ispace = strings.Index(restofdata, " ")
		if ispace < 0 {
			fmt.Println("Not enough arguments")
			return errors.New("Not enough arguments")
		}

		rolename = restofdata[:ispace]
		restofdata = restofdata[ispace+1:]

		roles, err := s.GuildRoles(m.GuildID)
		if err != nil {
			fmt.Println("Error getting roles.")
			return err
		}

		for _, role := range roles {
			if role.Name == rolename {
				roleID = role.ID
				break
			}
		}
	}

	if roleID == "" {
		fmt.Println("Did not find role in roles")
		return errors.New("Did not find role:" + rolename + ", in guild.")
	}

	var allow bool
	valuestring := strings.TrimSpace(restofdata)
	switch valuestring {
	case "yes":
		allow = true
	case "no":
		allow = false
	default:
		fmt.Println("Need yes or no")
		return errors.New("Need yes or no")
	}

	node.AddOrSetRolePerm(m.GuildID, roleID, allow)
	discord.masterconfig.Write()

	s.ChannelMessageSend(m.ChannelID, "Added value:'"+valuestring+"' to permission node:'"+node.FullName()+"', for role: <@&"+roleID+">")

	return nil
}

func (discord *DiscordHandler) handleAddPermUser(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	ispace := strings.Index(strings.TrimSpace(data), " ")
	nodepath := data[:ispace]
	restofdata := data[ispace+1:]

	node, err := discord.permhandler.GetPermNode(nodepath)
	if err != nil {
		fmt.Println("Node not found for path:", nodepath)
		return err
	}

	userID := ""
	// <@ %s >
	ibracket := strings.Index(strings.TrimSpace(restofdata), ">")
	if ibracket > 0 {
		rolemention := restofdata[:ibracket]
		restofdata = restofdata[ibracket+1:]
		userID = rolemention[2:len(rolemention)]
	} else {
		return errors.New("Need to mention user")
	}

	var allow bool
	valuestring := strings.TrimSpace(restofdata)
	switch valuestring {
	case "yes":
		allow = true
	case "no":
		allow = false
	default:
		fmt.Println("Need yes or no")
		return errors.New("Need yes or no")
	}

	node.AddOrSetUserPerm(m.GuildID, userID, allow)
	discord.masterconfig.Write()

	s.ChannelMessageSend(m.ChannelID, "Added value:'"+valuestring+"' to permission node:'"+node.FullName()+"', for user: <@"+userID+">")

	return nil
}

func (discord *DiscordHandler) handleRemovePermRole(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	ispace := strings.Index(strings.TrimSpace(data), " ")
	nodepath := data[:ispace]
	restofdata := data[ispace+1:]

	node, err := discord.permhandler.GetPermNode(nodepath)
	if err != nil {
		fmt.Println("Node not found for path:", nodepath)
		return err
	}

	roleID := ""
	rolename := ""
	// <@& %s >
	restofdata = strings.TrimSpace(restofdata)
	ibracket := strings.Index(restofdata, ">")
	if ibracket > 0 {
		rolemention := restofdata[:ibracket]
		restofdata = restofdata[ibracket+1:]
		roleID = rolemention[3:len(rolemention)]
	} else {
		rolename = restofdata

		roles, err := s.GuildRoles(m.GuildID)
		if err != nil {
			fmt.Println("Error getting roles.")
			return err
		}

		for _, role := range roles {
			if role.Name == rolename {
				roleID = role.ID
				break
			}
		}
	}

	if roleID == "" {
		fmt.Println("Did not find role in roles")
		return errors.New("Did not find role:" + rolename + ", in guild.")
	}

	node.RemoveRolePerm(m.GuildID, roleID)
	discord.masterconfig.Write()

	s.ChannelMessageSend(m.ChannelID, "Removed role: <@&"+roleID+"> from permission node:'"+node.FullName()+"'")

	return nil
}

func (discord *DiscordHandler) handleRemovePermUser(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	ispace := strings.Index(strings.TrimSpace(data), " ")
	nodepath := data[:ispace]
	restofdata := data[ispace+1:]

	node, err := discord.permhandler.GetPermNode(nodepath)
	if err != nil {
		fmt.Println("Node not found for path:", nodepath)
		return err
	}

	userID := ""
	// <@ %s >
	ibracket := strings.Index(strings.TrimSpace(restofdata), ">")
	if ibracket > 0 {
		rolemention := restofdata[:ibracket]
		restofdata = restofdata[ibracket+1:]
		userID = rolemention[2:len(rolemention)]
	} else {
		return errors.New("Need to mention user")
	}

	node.RemoveUserPerm(m.GuildID, userID)
	discord.masterconfig.Write()

	s.ChannelMessageSend(m.ChannelID, "Removed user: <@"+userID+"> from permission node:'"+node.FullName()+"'")

	return nil
}

func (discord *DiscordHandler) handleListPermissionsForNode(data string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	var permFields []*discordgo.MessageEmbedField

	node, err := discord.permhandler.GetPermNode(strings.TrimSpace(data))
	if err != nil {
		fmt.Println("Unable to find node:", data)
		return err
	}

	guildperm, err := node.GetGuildPerm(m.GuildID)
	if err != nil {
		fmt.Println("No guildperms for guild:", data)
	} else {

		// Users
		for _, perm := range guildperm.PermsUser() {
			usernick, err := s.GuildMember(m.GuildID, perm.PermID())
			if err != nil {
				fmt.Println("Error fetching GuildMember:", perm.PermID())
				continue
			}

			nametoshow := usernick.Nick
			if nametoshow == "" {
				nametoshow = usernick.User.Username
			}

			permFields = append(permFields, &discordgo.MessageEmbedField{
				Name:   nametoshow + "\t" + strconv.FormatBool(perm.PermAllowed()),
				Value:  "User",
				Inline: false,
			})
		}

		// Roles
		roles, err := s.GuildRoles(m.GuildID)
		if err != nil {
			fmt.Println("Error fetching GuildRoles")
			return err
		}
		for _, perm := range guildperm.PermsRole() {
			var foundRole *discordgo.Role
			for _, r := range roles {
				if r.ID == perm.PermID() {
					foundRole = r
				}
			}

			if foundRole == nil {
				fmt.Println("Error finding role:", perm.PermID())
				continue
			}

			permFields = append(permFields, &discordgo.MessageEmbedField{
				Name:   foundRole.Name + "\t" + strconv.FormatBool(perm.PermAllowed()),
				Value:  "Role",
				Inline: false,
			})
		}
	}

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00,
		Description: node.FullName(),
		Fields:      permFields,
		Timestamp:   time.Now().Format(time.RFC3339),
		Title:       node.Name(),
	}

	_, err = s.ChannelMessageSendEmbed(m.Message.ChannelID, embed)
	if err != nil {
		fmt.Println("Error embedding message", err)
		return err
	}
	return nil
}
