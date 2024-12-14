package main

import (
	"fmt"
	"goantisc/internal/core"
	"goantisc/internal/data"
	"goantisc/internal/handler"
	"goantisc/internal/helper/checker"
	"goantisc/internal/services"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goantisc/internal/log"

	"github.com/longbridgeapp/opencc"
	tele "gopkg.in/telebot.v4"
)

func main() {
	// Loading config.
	servCfg, err := core.LoadConfig("config.toml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Initialize logger.
	log.InitLogger()
	defer log.DisposeLogger()

	// Intiialize dictionary.
	dic, err := opencc.New("s2t")
	if err != nil {
		log.Logger().Fatalf(err.Error())
	}
	if err != nil {
		log.Logger().Fatalf(err.Error())
	}
	_ = dic

	// Initialize the store.
	chatStore := data.NewChatInfoStore()
	if err := chatStore.Load("chatinfo.json"); err != nil {
		if err := chatStore.Save("chatinfo.json"); err != nil {
			log.Logger().Fatal(err.Error())
		}
	}
	log.Logger().Infof("[Goantib] Loading saved data successful")

	log.Logger().Infof("[Goantib] Try to connect telegram api....")
	// Initialize bot.
	bot, err := tele.NewBot(tele.Settings{
		Token:  servCfg.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})

	scChecker := checker.NewSimplifiedChineseChecker(dic)
	delWatcher := services.NewDeleteWatcher(bot)
	delWatcher.Start()

	// Mount the handler.
	handler.MountGroupMessageHandler(bot, scChecker, delWatcher)

	// Start to polling the data.
	go bot.Start()
	log.Logger().Infof("[Goantib] Start to polling message")

	// Handle signal.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	<-c
	fmt.Println("\n")
	delWatcher.Stop()
	bot.Stop()
	log.Logger().Infof("[Goantib] Shutdown successfully.")
}

func initBotHandler(bot *tele.Bot) {
	bot.Handle("/start", func(ctx tele.Context) error {
		ctx.Reply("Hi")
		return nil
	})
}
