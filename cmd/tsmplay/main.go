package main

import (
	"fmt"
	"github.com/Muges/tsm/streamer"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"os"
	"time"
)

func main() {
	// Read command-line arguments
	if len(os.Args) != 2 {
		fmt.Println("usage: tsmplay filename.wav")
		os.Exit(1)
	}
	filename := os.Args[1]

	// Open and decode wav file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("error: unable to open file \"%s\"\n", filename)
		fmt.Println(err)
		os.Exit(1)
	}
	stream, format, err := wav.Decode(file)
	if err != nil {
		fmt.Printf("error: \"%s\" is not a valid wav file\n", filename)
		fmt.Println(err)
		os.Exit(1)
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Create a channel that will be closed at the end of playback
	done := make(chan struct{})

	stretched_stream := streamer.Streamer{Streamer: stream}

	speaker.Play(beep.Seq(stretched_stream, beep.Callback(func() {
		close(done)
	})))

	// Wait for the channel to be closed before quitting
	<-done
}
