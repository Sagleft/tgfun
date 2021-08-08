package main

import (
	"log"
	"os"

	tgfun "./lib"
)

func main() {
	// get env params
	botToken := os.Getenv("TOKEN")

	data := tgfun.FunnelData{
		Token: botToken,
	}
	script := tgfun.FunnelScript{
		"/start": tgfun.FunnelEvent{
			Message: tgfun.EventMessage{
				Text:  "Welcome message text",
				Image: "image.jpg",
				Buttons: []tgfun.MessageButton{
					tgfun.MessageButton{
						Text:          "next",
						NextMessageID: "nextmsg",
					},
					tgfun.MessageButton{
						Text: "test url button",
						URL:  "https://example.com",
					},
				},
			},
		},
		"nextmsg": tgfun.FunnelEvent{
			Message: tgfun.EventMessage{
				Text: "Next message text",
			},
		},
	}

	/*container := struct {
		Data   tgfun.FunnelData   `json:"data"`
		Script tgfun.FunnelScript `json:"script"`
	}{
		Data:   data,
		Script: script,
	}

	bytes, _ := json.Marshal(container)
	fmt.Println(string(bytes))
	fmt.Println()*/

	// create funnel
	funnel, err := tgfun.NewFunnel(data, script)
	// check error
	if err != nil {
		log.Fatalln("failed to create funnel: " + err.Error())
	}

	// run funnel
	err = funnel.Run()
	if err != nil {
		log.Fatalln("failed to run funnel: " + err.Error())
	}

	printFunnelArt(true)

	forever := make(chan bool)
	// run in background
	<-forever
}
