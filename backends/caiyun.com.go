package backends

import (
    "encoding/json"
    "flag"
    "fmt"
    "github.com/kernelhcy/wego/iface"
    "io/ioutil"
    "log"
    "net/http"
    "time"
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
    Chn float32	`json:"chn"`
    Usa float32	`json:"usa"`
}

type cyAqiDescription struct {
    Chn string	`json:"chn"`
    Usa string	`json:"usa"`
}

type cyRealtimeAirQuality struct {
    Pm25 float32          `json:"pm25"`
    Pm10 float32          `json:"pm10"`
    O3   float32          `json:"o3"`
    SO2  float32          `json:"so2"`
    NO2  float32          `json:"no2"`
    CO   float32          `json:"co"`
    Aqi  cyAqi            `json:"aqi"`
    Desc cyAqiDescription `json:"description"`
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
    Aqi 				cyRealtimeAirQuality	`json:"air_quality"`
    LifeIndex 			cyLifeIndices			`json:"life_index"`
}

type cyTimeValue struct {
    DateTime    time.Time   `json:"datetime"`
}

func (v *cyTimeValue) UnmarshalJSON(data []byte) (err error) {
    vv := &struct { DateTime   string `json:"datetime"` }{""}
    err = json.Unmarshal(data, vv)
    if err != nil {
        return err
    }

    v.DateTime, err = time.Parse("2006-01-02T15:04-07:00", vv.DateTime)
    return err
}

type cyTimeIntValue struct {
    cyTimeValue
    Value       int32       `json:"value"`
}

func (v *cyTimeIntValue) UnmarshalJSON(data []byte) (err error) {
    err = v.cyTimeValue.UnmarshalJSON(data)
    if err != nil {
        return err
    }

    vv := &struct { Value int32 `json:"value"` }{0}
    err = json.Unmarshal(data, vv)
    v.Value = vv.Value
    return err
}

type cyTimeFloatValue struct {
    cyTimeValue
    Value       float32     `json:"value"`
}

func (v *cyTimeFloatValue) UnmarshalJSON(data []byte) (err error) {
    err = v.cyTimeValue.UnmarshalJSON(data)
    if err != nil {
        return err
    }

    vv := &struct { Value float32 `json:"value"` }{0.0}
    err = json.Unmarshal(data, vv)
    v.Value = vv.Value
    return err
}

type cyTimeStringValue struct {
    cyTimeValue
    Value       string      `json:"value"`
}

func (v *cyTimeStringValue) UnmarshalJSON(data []byte) (err error) {
    err = v.cyTimeValue.UnmarshalJSON(data)
    if err != nil {
        return err
    }

    vv := &struct { Value string `json:"value"` }{""}
    err = json.Unmarshal(data, vv)
    v.Value = vv.Value
    return err
}

type cyHourly struct {
    Status      string              `json:"status"`
    Desc        string              `json:"description"`
    Temperature []cyTimeFloatValue  `json:"temperature"`
    Humidity    []cyTimeFloatValue  `json:"humidity"`
    Pressure    []cyTimeFloatValue  `json:"pressure"`
    Visibility  []cyTimeFloatValue  `json:"visibility"`
    Dswrf       []cyTimeFloatValue  `json:"dswrf"`
    Skycon      []cyTimeStringValue `json:"skycon"`
}

type cyWeatherResult struct {
    Realtime    cyRealtime	`json:"realtime"`
    Hourly      cyHourly    `json:"hourly"`
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

    fmt.Printf("%+v\n", data)

    return iface.Data{}
}

func init() {
    iface.AllBackends["caiyun.com"] = &cyConfig{}
}
