package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type IpifyApistruct struct {
	IP string `json:"ip"`
}

type Latlonapistruct struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	City      string  `json:"city"`
}

type Weatherinfostruct struct {
	Current struct {
		Temperature   float64 `json:"temperature_2m"`
		WindSpeed     float64 `json:"wind_speed_10m"`
		WindGusts     float64 `json:"wind_gusts_10m"`
		Precipitation float64 `json:"precipitation"`
		Rain          float64 `json:"rain"`
		Snowfall      float64 `json:"snowfall"`
	} `json:"current"`
}

type Ollamarequeststruct struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Format string `json:"format"`
	Stream bool   `json:"stream"`
}

type Ollamaresponsestruct struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func main() {
	fmt.Println("Ollama LLM weather IP location checker!")
	// get the users ip with ipify API
	req, _ := http.Get("https://api.ipify.org?format=json")
	defer req.Body.Close()
	ioreq, _ := ioutil.ReadAll(req.Body)
	var Ipify IpifyApistruct
	err := json.Unmarshal(ioreq, &Ipify)
	if err != nil {
		fmt.Println(err)
	}
	// ok user ip stored now as Ipify.IP now we get latitude and longitude from dat
	latlongetreq := string("http://ip-api.com/json/" + Ipify.IP)
	req, _ = http.Get(latlongetreq)
	defer req.Body.Close()
	ioreq, _ = ioutil.ReadAll(req.Body)
	var latlon Latlonapistruct
	err = json.Unmarshal(ioreq, &latlon)
	if err != nil {
		fmt.Println(err)
	}
	// we now have user latitude and longitude as latlon.Latitude and latlon.Longitude, now we get openmeteo weather info to pump into the llm
	weatherRequest := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,wind_speed_10m,wind_gusts_10m,precipitation,rain,snowfall",
		latlon.Latitude,
		latlon.Longitude,
	)
	req, _ = http.Get(weatherRequest)
	defer req.Body.Close()
	ioreq, _ = ioutil.ReadAll(req.Body)
	var WeatherInfo Weatherinfostruct
	err = json.Unmarshal(ioreq, &WeatherInfo)
	if err != nil {
		fmt.Println(err)
	}

	// ok now lets figure out http.post and ollama local api :/
	ollamaurl := "http://localhost:11434/api/generate"
	ollamaWeatherPrompt := fmt.Sprintf(
		"You are now WeatherReporterLLM, a charismatic and professional weather reporter delivering a live national broadcast. Your job is to report the current weather conditions in the given city with clarity, confidence, and a touch of warmth. Speak naturally, as if you're on camera, using up to 4 engaging sentences. just deliver the weather like a real person on TV. ONLY RESPOND WITH WHAT YOU WOULD SAY, NO JSON."+
			"Here is the raw weather data: Temperature in Celsius: %.1f, Wind speed in KMH: %.1f, Wind gust speed in KMH: %.1f, Precipitation in MM: %.1f, Rain amount: %.1f, Snowfall amount: %.1f, City: %s",
		WeatherInfo.Current.Temperature,
		WeatherInfo.Current.WindSpeed,
		WeatherInfo.Current.WindGusts,
		WeatherInfo.Current.Precipitation,
		WeatherInfo.Current.Rain,
		WeatherInfo.Current.Snowfall,
		latlon.City,
	)

	ollamareq := Ollamarequeststruct{
		Model:  "llama3.2",
		Prompt: ollamaWeatherPrompt,
		Format: "json",
		Stream: false,
	}
	Ollamajsondata, _ := json.Marshal(ollamareq)
	if err != nil {
		fmt.Println(err)
	}
	req, _ = http.Post(ollamaurl, "application/json", bytes.NewBuffer(Ollamajsondata))

	ollamaIO, _ := ioutil.ReadAll(req.Body)

	var ollamaResp Ollamaresponsestruct
	err = json.Unmarshal(ollamaIO, &ollamaResp)
	if err != nil {
		fmt.Println("Failed to parse response:", err)
		fmt.Println(string(ollamaIO))
		return
	}

	// one small issue here, response prints with {""} around it, cant tell if its the llm or just my ahh coding skillz
	fmt.Println(ollamaResp.Response)

}
