package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"github.com/rs/zerolog/log"
)

func main() {
	api, err := maxbot.New(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed. Stop.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGTERM, os.Interrupt)
		<-exit
		cancel()
	}()

	info, err := api.Bots.GetBot(ctx)
	log.Printf("Get me: %#v %#v", info, err)

	for upd := range api.GetUpdates(ctx) {
		api.Debugs.Send(ctx, upd)
		switch upd := upd.(type) {
		case *schemes.MessageCreatedUpdate:
			handleMessageCreated(api, ctx, upd)
		case *schemes.MessageCallbackUpdate:
			handleCallback(api, ctx, upd)
		default:
			log.Printf("Unknown type: %#v", upd)
		}
	}
}
