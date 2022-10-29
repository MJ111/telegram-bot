package main

import (
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	// load .env file from given path
	// we keep it empty it will load .env from current directory
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))

	var chatid int64
	chatid, _ = strconv.ParseInt(os.Getenv("CHATROOM_ID"), 10, 64)

	if err != nil {
		panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	ticker := time.NewTicker(30 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				msg2 := tgbotapi.NewMessage(chatid, searchYoutube("jazz guitar"))
				if _, err := bot.Send(msg2); err != nil {
					//log.Panic(err)
					fmt.Printf("%s", err)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "help":
			msg.Text = "youtubes: /jjjj, /wwww"
		case "jjjj":
			msg.Text = searchYoutube("jazz guitar")
		case "wwww":
			msg.Text = searchYoutube("muscle workout")
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := bot.Send(msg); err != nil {
			//log.Panic(err)
		}
	}
}

var (
	maxResults = flag.Int64("max-results", 1, "Max YouTube results")
)

func searchYoutube(inputQuery string) string {
	flag.Parse()

	client := &http.Client{
		Transport: &transport.APIKey{Key: os.Getenv("GOOGLE_API_KEY")},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// Make the API call to YouTube.
	call := service.Search.List([]string{"id", "snippet"}).
		Q(inputQuery).
		MaxResults(*maxResults).
		Order("date")

	response, err := call.Do()
	//handleError(err, "")

	// Group video, channel, and playlist results in separate lists.
	videos := make(map[string]string)
	//channels := make(map[string]string)
	//playlists := make(map[string]string)

	var videoID string

	// Iterate through each item and add it to the correct list.
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			videos[item.Id.VideoId] = item.Snippet.Title
			videoID = item.Id.VideoId
			//case "youtube#channel":
			//	channels[item.Id.ChannelId] = item.Snippet.Title
			//case "youtube#playlist":
			//	playlists[item.Id.PlaylistId] = item.Snippet.Title
		}
	}

	//printIDs("Videos", videos)
	//printIDs("Channels", channels)
	//printIDs("Playlists", playlists)

	return fmt.Sprintf("https://www.youtube.com/watch?v=%v", videoID)
}

// Print the ID and title of each result in a list as well as a name that
// identifies the list. For example, print the word section name "Videos"
// above a list of video search results, followed by the video ID and title
// of each matching video.
func printIDs(sectionName string, matches map[string]string) {
	fmt.Printf("%v:\n", sectionName)
	for id, title := range matches {
		fmt.Printf("[https://www.youtube.com/watch?v=%v] %v\n", id, title)
	}
	fmt.Printf("\n\n")
}
