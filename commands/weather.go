package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"luna/interfaces"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

type weatherResponse struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Name string `json:"name"`
}

type WeatherCommand struct {
	APIKey string
	Log    interfaces.Logger
}

func (c *WeatherCommand) GetCommandDef() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "weather",
		Description: "指定した都市の現在の天気を表示します（実験的機能）(APIが設定されていないので利用できません)",
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "city", Description: "都市名 (例: Tokyo)", Required: true},
		},
	}
}

func (c *WeatherCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if c.APIKey == "" {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ 天気機能は現在利用できません。", Flags: discordgo.MessageFlagsEphemeral},
		}); err != nil {
			c.Log.Error("Failed to respond to interaction", "error", err)
		}
		return
	}

	city := i.ApplicationCommandData().Options[0].StringValue()
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric&lang=ja", city, c.APIKey)

	resp, err := http.Get(url)
	if err != nil {
		c.Log.Error("天気情報の取得に失敗", "error", err, "city", city)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ 都市が見つかりませんでした。", Flags: discordgo.MessageFlagsEphemeral},
		}); err != nil {
			c.Log.Error("Failed to respond to interaction", "error", err)
		}
		return
	}

	body, _ := io.ReadAll(resp.Body)
	var data weatherResponse
	if err := json.Unmarshal(body, &data); err != nil {
		c.Log.Error("Failed to unmarshal weather response", "error", err)
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s の天気", data.Name),
		Color: 0x42b0f4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "天気", Value: data.Weather[0].Description, Inline: true},
			{Name: "気温", Value: fmt.Sprintf("%.1f °C", data.Main.Temp), Inline: true},
			{Name: "湿度", Value: fmt.Sprintf("%d %%", data.Main.Humidity), Inline: true},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: fmt.Sprintf("http://openweathermap.org/img/wn/%s@2x.png", data.Weather[0].Icon)},
	}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
	}); err != nil {
		c.Log.Error("Failed to send weather response", "error", err)
	}
}

func (c *WeatherCommand) HandleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {}
func (c *WeatherCommand) HandleModal(s *discordgo.Session, i *discordgo.InteractionCreate)     {}
func (c *WeatherCommand) GetComponentIDs() []string                                            { return []string{} }
func (c *WeatherCommand) GetCategory() string {
	return "ユーティリティ"
}