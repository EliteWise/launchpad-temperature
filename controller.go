package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/rakyll/portmidi"
)

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

func main() {
	portmidi.Initialize()
	defer portmidi.Terminate()

	deviceID := -1
	for i := 0; i < portmidi.CountDevices(); i++ {
		info := portmidi.Info(portmidi.DeviceID(i))
		fmt.Println(info.Name)
		if info.IsOutputAvailable && info.Name == "Launchpad MK2 MIDI 1" {
			deviceID = i
			break
		}
	}

	if deviceID == -1 {
		fmt.Println("Launchpad non trouvé.")
		return
	}

	out, err := portmidi.NewOutputStream(portmidi.DeviceID(deviceID), 1024, 0)
	if err != nil {
		fmt.Println("Erreur ouverture du port MIDI:", err)
		return
	}
	defer out.Close()

	temp := getTemperature("Montpellier")
	fmt.Println(temp)
}
