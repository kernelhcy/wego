package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kernelhcy/wego/iface"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	caiyunUri="https://api.caiyunapp.com/v2.5/%s/%f,%f/weather.json?alert=true"
)

type cyConfig struct {
	token string
	latitude  float64
	longitude float64
}

type cyLifeIndex struct {
	Index 	float32	`json:"index"`
	Desc 	string	`json:"desc"`
}

type cyLifeIndices struct {
	Ultraviolet cyLifeIndex	`json:"ultraviolet"`
	Comfort 	cyLifeIndex	`json:"comfort"`
}

type cyAqi struct {
	Chn int32	`json:"chn"`
	Usa int32	`json:"usa"`
}

type cyAirDescription struct {
	Chn string	`json:"chn"`
	Usa string	`json:"usa"`
}

type cyRealtimeAirQuality struct {
	Pm25 int32            `json:"pm25"`
	Pm10 int32            `json:"pm10"`
	O3   int32            `json:"o3"`
	SO2  int32            `json:"so2"`
	NO2  int32            `json:"no2"`
	CO   float32          `json:"co"`
	Aqi  cyAqi            `json:"aqi"`
	Desc cyAirDescription `json:"description"`
}

type cyRealtimeWind struct {
	Speed 		float32	`json:"speed"`
	Direction 	float32	`json:"direction"`
}

type cyRealtime struct {
	Status 				string					`json:"status"`
	Temperature 		float32					`json:"temperature"`
	Humidity 			float32					`json:"humidity"`
	CloudRate 			float32					`json:"cloudrate"`
	Skycon 				string					`json:"skycon"`
	Visibility 			float32					`json:"visibility"`
	Dswrf 				float32					`json:"dswrf"`
	Wind 				cyRealtimeWind			`json:"wind"`
	Pressure 			float32					`json:"pressure"`
	ApparentTemperature float32					`json:"apparent_temperature"`
	Aqi 				cyRealtimeAirQuality	`json:"aqi_quality"`
	LifeIndex 			cyLifeIndices			`json:"life_index"`
}

type cyWeatherResult struct {
	Realtime cyRealtime	`json:"realtime"`
}

type cyWeatherData struct {
	Status 		string				`json:"status"`
	ApiVersion 	string				`json:"api_version"`
	ApiStatus 	string				`json:"api_status"`
	Lang 		string				`json:"lang"`
	Unit 		string				`json:"unit"`
	Tzshift 	int32				`json:"tzshift"`
	Timezone 	string				`json:"timezone"`
	ServerTime 	int32				`json:"server_time"`
	Location 	[2]float32			`json:"location"`
	Result 		cyWeatherResult		`json:"result"`
}

func (c *cyConfig) fetch() (*cyWeatherData, error) {
	url := fmt.Sprintf(caiyunUri, c.token, c.longitude, c.latitude)
	log.Printf("caiyun url: %s\n", url)
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to get (%s): %v", url, err)
	} else if res.StatusCode != 200 {
		return nil, fmt.Errorf("unable to get (%s): http status %d", url, res.StatusCode)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body (%s): %v", url, err)
	}

	var data cyWeatherData
	if err = json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response (%s): %v\nThe json body is: %s", url, err, string(body))
	}

	return &data, nil
}

func (c *cyConfig) Setup() {
	flag.StringVar(&c.token, "caiyun-api-token", "", "caiyun backend: the api `TOKEN` to use")
	flag.Float64Var(&c.latitude, "caiyun-latitude", 30.274085, "caiyun backend: the `latitude` to request from caiyun.com")
	flag.Float64Var(&c.longitude, "caiyun-longitude", 120.15507, "caiyun backend: the `longitude` to request from caiyun.com")
}

func (c *cyConfig) Fetch(location string, numdays int) iface.Data {
	data, err := c.fetch()
	if err != nil {
		log.Printf("caiyun fetch error: %v\n", err)
		return iface.Data{}
	}

	fmt.Printf("%s %d %s\n", data.ApiVersion, data.ServerTime, data.Result.Realtime.Skycon)

	return iface.Data{}
}

func init() {
	iface.AllBackends["caiyun.com"] = &cyConfig{}
}
