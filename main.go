package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	// "time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Images struct {
	Small string `json:"small"`
	Large string `json:"large"`
}

type CurrentData struct {
	CurrentTier          int    `json:"currenttier"`
	CurrentTierPatched   string `json:"currenttierpatched"`
	Images               Images `json:"images"`
	RankingInTier        int    `json:"ranking_in_tier"`
	MMRChangeLastGame    int    `json:"mmr_change_to_last_game"`
	ELO                  int    `json:"elo"`
	GamesNeededForRating int    `json:"games_needed_for_rating"`
}

type Data struct {
	Name        string      `json:"name"`
	Tag         string      `json:"tag"`
	PUUID       string      `json:"puuid"`
	CurrentData CurrentData `json:"current_data"`
}

type Response struct {
	Data Data `json:"data"`
}

type PlaylistData struct {
	Playlist     string `json:"playlist"`
	MMR          int    `json:"mmr"`
	Rank         int    `json:"rank"`
	Division     int    `json:"division"`
	WinStreak    string `json:"win_streak"`
	RankName     string `json:"rankName"`
	DivisionName string `json:"divisionName"`
	DeltaUp      int    `json:"deltaUp"`
	DeltaDown    int    `json:"deltaDown"`
}

type RLResponse struct {
	DisplayName string         `json:"displayName"`
	Rankings    []PlaylistData `json:"rankings"`
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = goDotEnvVariable("TOKEN")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func rankToColour(rank string) int {

	switch rank {
	case "Unranked":
		return 0x797373
	case "Iron":
		return 0x3b3b3b
	case "Bronze":
		return 0x69450d
	case "Silver":
		return 0xbbbfbe
	case "Gold":
		return 0xdd9623
	case "Platinum":
		return 0x328d9e
	case "Diamond":
		return 0xd781e9
	case "Ascendant":
		return 0x1e8a51
	case "Immortal":
		return 0xb02639
	case "Radiant":
		return 0xfce29b
	default:
		return 0xFFFFFF
	}

}

func getDelta(deltaUp int, deltaDown int) string {

	if deltaUp > deltaDown {
		return fmt.Sprintf(" ▼ %d", deltaDown)
	} else if deltaUp < deltaDown {
		return fmt.Sprintf(" ▲ %d", deltaUp)
	} else {
		return ""
	}

}

var (
	commands = []*discordgo.ApplicationCommand{

		{
			Name:        "rlrank",
			Description: "Get users Rocket League rank",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "Username",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "platform",
					Description: "Platform",
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Steam",
							Value: "steam",
						},
						{
							Name:  "Epic",
							Value: "epic",
						},
						{
							Name:  "Xbox",
							Value: "xbl",
						},
						{
							Name:  "Playstation",
							Value: "psn",
						},
					},

					Required: true,
				},
			},
		},
		{
			Name:        "valrank",
			Description: "Get users Valorant rank",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "Username",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "tag",
					Description: "Tag after #",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "region",
					Description: "Region",
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "North America/South America",
							Value: "na",
						},
						{
							Name:  "Europe",
							Value: "eu",
						},
						{
							Name:  "Asia-Pacific",
							Value: "ap",
						},
						{
							Name:  "Korea",
							Value: "kr",
						},
					},
					Required: true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){

		"rlrank": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Access options in the order provided by the user.
			options := i.ApplicationCommandData().Options

			// Or convert the slice into a map
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Printf("Error sending deferred response: %v", err)
			}

			// get the value from the option map

			platform, ok := optionMap["platform"]

			if !ok {
				log.Fatalf("Error getting platform")
			}

			username, ok := optionMap["username"]

			if !ok {
				log.Fatalf("Error getting username")
			}

			resp, err := http.Get(fmt.Sprintf(goDotEnvVariable("RL_API")+"/%s/%s?raw=true", platform.StringValue(), username.StringValue()))
			if err != nil {
				log.Fatalln(err)
			}

			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)

			var rlResponse RLResponse

			err = json.Unmarshal(body, &rlResponse)
			var message string

			for _, rank := range rlResponse.Rankings {
				if rank.Playlist == "unranked" {
					continue
				}
				message += fmt.Sprintf("%s: %s %s (%d%s)\n", strings.Title(strings.Replace(strings.ReplaceAll(rank.Playlist, "_", " "), "v", "V", 1)), rank.RankName, rank.DivisionName, rank.MMR, getDelta(rank.DeltaUp, rank.DeltaDown))

			}

			if err != nil {
				log.Fatalln(err)

				_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Embeds: &[]*discordgo.MessageEmbed{{
						Title: "Failed to get rank for " + username.StringValue(),
					}},
				})

			}

			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{{
					Title:       "Ranks for " + rlResponse.DisplayName,
					Description: fmt.Sprintf(message),
				}},
			})

			if err != nil {
				log.Printf("Error editing response: %v", err)
			}
		},
		"valrank": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

			options := i.ApplicationCommandData().Options

			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Printf("Error sending deferred response: %v", err)
			}

			// get the value from the option map

			region, ok := optionMap["region"]

			if !ok {
				log.Fatalf("Error getting region")
			}

			username, ok := optionMap["username"]

			if !ok {
				log.Fatalf("Error getting username")
			}

			tag, ok := optionMap["tag"]

			if !ok {
				log.Fatalf("Error getting tag")
			}

			resp, err := http.Get(fmt.Sprintf(goDotEnvVariable("VAL_API")+"/%s/%s/%s", region.StringValue(), username.StringValue(), tag.StringValue()))
			if err != nil {
				log.Fatalln(err)
			}

			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)

			// convert body to object
			var responseData Response
			error := json.Unmarshal([]byte(body), &responseData)
			if error != nil {
				log.Fatalln(error)
				_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Embeds: &[]*discordgo.MessageEmbed{{
						Title: "Failed to get ranks for " + username.StringValue() + "#" + tag.StringValue(),
					}},
				})
				return

			}

			// check if the user exists

			if responseData.Data.Name == "" {
				_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Embeds: &[]*discordgo.MessageEmbed{{
						Title: "Failed to get ranks for " + username.StringValue() + "#" + tag.StringValue(),
					}},
				})
				return
			}

			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{{
					Title:       "Rank for " + string(responseData.Data.Name) + "#" + string(responseData.Data.Tag),
					Description: fmt.Sprintf("Rank: %s\nELO: %d\nMMR Last Game: %d", responseData.Data.CurrentData.CurrentTierPatched, responseData.Data.CurrentData.ELO, responseData.Data.CurrentData.MMRChangeLastGame),
					// Image: &discordgo.MessageEmbedImage{
					// 	URL: string(responseData.Data.CurrentData.Images.Large),
					// },
					Color: rankToColour(strings.Split(responseData.Data.CurrentData.CurrentTierPatched, " ")[0]),
				}},
			})

			if err != nil {
				log.Printf("Error editing response: %v", err)
			}
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")
		// // We need to fetch the commands, since deleting requires the command ID.
		// // We are doing this from the returned commands on line 375, because using
		// // this will delete all the commands, which might not be desirable, so we
		// // are deleting only the commands that we added.
		// registeredCommands, err := s.ApplicationCommands(s.State.User.ID, *GuildID)
		// if err != nil {
		// 	log.Fatalf("Could not fetch registered commands: %v", err)
		// }

		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
