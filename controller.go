package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/rakyll/portmidi"
)

// 5 l x 3 col
var digits = map[rune][5][3]bool{
	'0': {
		{true, true, true},
		{true, false, true},
		{true, false, true},
		{true, false, true},
		{true, true, true},
	},
	'1': {
		{false, true, false},
		{false, true, false},
		{false, true, false},
		{false, true, false},
		{false, true, false},
	},
	'2': {
		{true, true, true},
		{false, false, true},
		{true, true, true},
		{true, false, false},
		{true, true, true},
	},
	'3': {
		{true, true, true},
		{false, false, true},
		{true, true, true},
		{false, false, true},
		{true, true, true},
	},
	'4': {
		{true, false, true},
		{true, false, true},
		{true, true, true},
		{false, false, true},
		{false, false, true},
	},
	'5': {
		{true, true, true},
		{true, false, false},
		{true, true, true},
		{false, false, true},
		{true, true, true},
	},
	'6': {
		{true, true, true},
		{true, false, false},
		{true, true, true},
		{true, false, true},
		{true, true, true},
	},
	'7': {
		{true, true, true},
		{false, false, true},
		{false, true, false},
		{true, false, false},
		{true, false, false},
	},
	'8': {
		{true, true, true},
		{true, false, true},
		{true, true, true},
		{true, false, true},
		{true, true, true},
	},
	'9': {
		{true, true, true},
		{true, false, true},
		{true, true, true},
		{false, false, true},
		{true, true, true},
	},
}

func getTemperature(city string) string {
	requestUrl := fmt.Sprintf("http://wttr.in/%s?format=%%t", city)
	res, err := http.Get(requestUrl)
	if err != nil {
		fmt.Printf("Error making http request: %s\n", err)
	}
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
	}

	re := regexp.MustCompile("[0-9]+")
	return strings.TrimSpace(re.FindAllString(string(respBody), -1)[0])
}

func sendMidiMessage(stream *portmidi.Stream, status, data1, data2 int64) {
	stream.WriteShort(status, data1, data2)
}

func displayDigit(stream *portmidi.Stream, digit rune, xOffset int) {
	pattern, exists := digits[digit]
	if !exists {
		return
	}

	for row := 0; row < 5; row++ {
		for col := 0; col < 3; col++ {
			if pattern[row][col] {
				note := 81 - (row * 10) + col + (xOffset * 4)
				sendMidiMessage(stream, 0x90, int64(note), 5)
			}
		}
	}
}

func displayTemperature(stream *portmidi.Stream, temp string) {
	xOffset := 0
	for _, digit := range temp {
		displayDigit(stream, digit, xOffset)
		xOffset += 1
	}
}

func clearLaunchpad(stream *portmidi.Stream) {
	for note := 11; note <= 88; note++ {
		sendMidiMessage(stream, 0x90, int64(note), 0)
	}
}

func main() {
	city := "Paris"
	if len(os.Args) > 1 {
		city = os.Args[1]
	}

	portmidi.Initialize()
	defer portmidi.Terminate()

	deviceID := -1
	for i := 0; i < portmidi.CountDevices(); i++ {
		info := portmidi.Info(portmidi.DeviceID(i))
		if info.IsOutputAvailable && strings.Contains(info.Name, "Launchpad") {
			deviceID = i
			break
		}
	}

	if deviceID == -1 {
		fmt.Println("Launchpad not found.")
		return
	}

	out, err := portmidi.NewOutputStream(portmidi.DeviceID(deviceID), 1024, 0)
	if err != nil {
		fmt.Println("Error opening MIDI port.", err)
		return
	}
	defer out.Close()

	temp := getTemperature(city)
	fmt.Println(temp)

	clearLaunchpad(out)
	displayTemperature(out, temp)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	clearLaunchpad(out)
}
