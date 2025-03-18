package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "math"
    "net/http"
    "strings"
    "time"

    "github.com/shirou/gopsutil/v4/disk"
    "github.com/shirou/gopsutil/v4/sensors"
    "github.com/shirou/gopsutil/v4/host"
)

// APIResponse - define structure of api response
type APIResponse struct {
    Code    int         `json:"code"`
    Status  string      `json:"status"`
    Data    interface{} `json:"data"`
}

// api status enum
var API_MSG_ENUM = map[int]string{
    0:    "success",
    404:  "no data",
    522:  "wrong arguments",
}

// makeJSONResponse - format the response
func makeJSONResponse(code int, data interface{}) APIResponse {
    if _, ok := API_MSG_ENUM[code];!ok {
        panic("illegal code")
    }
    return APIResponse{
        Code:    code,
        Status:  API_MSG_ENUM[code],
        Data:    data,
    }
}

// roundToNDecimal 
func roundToNDecimal(num float64, n int) float64 {
    multiplier := math.Pow(10, float64(n))
    return math.Round(num*multiplier) / multiplier
}

// getDU - disk usage of linux root partition
func getDU() (interface{}, error) {
    usage, err := disk.Usage("/")
    if err != nil {
        log.Printf("Error getting disk usage: %v", err)
        return nil, err
    }
    return map[string]float64{
        "free_size": roundToNDecimal(float64(usage.Free)/(1024*1024), 2),
        "free_rate": roundToNDecimal(100-float64(usage.UsedPercent), 2),
    }, nil
}

// getTemps - read temperatures of linux sensors
func getTemps() (interface{}, error) {
    temps, err := sensors.SensorsTemperatures()
    if err != nil {
        return nil, err
    }
    result := make(map[string]map[string]float64)
    for _, temp := range temps {
        result[temp.SensorKey] = map[string]float64{
            "curr": roundToNDecimal(temp.Temperature, 2),
            "crit": roundToNDecimal(temp.Critical, 2),
        }
    }
    return result, nil
}

// getBootTime - sys uptime
func getBootTime() (interface{}, error) {
    bootTime, err := host.BootTime()
    if err != nil {
        return nil, err
    }
    bootTimeStamp := time.Unix(int64(bootTime), 0)
    elapsedSeconds := time.Since(bootTimeStamp).Seconds()
    return map[string]interface{}{
        "boot_time_str":    bootTimeStamp.Format("2006-01-02 15:04:05"),
        "boot_timestamp":   bootTimeStamp.Unix(),
        "elapsed_seconds":  roundToNDecimal(elapsedSeconds, 2),
        "elapsed_readable": formatElapsedTime(elapsedSeconds),
    }, nil
}

// formatElapsedTime - to get readable time format
func formatElapsedTime(seconds float64) string {
    var intervals = []struct {
        name  string
        count float64
    }{
        {"w", 604800},
        {"d", 86400},
        {"h", 3600},
        {"m", 60},
        {"s", 1},
    }
    result := []string{}
    for _, interval := range intervals {
        value := int(seconds / interval.count)
        if value > 0 {
            seconds -= float64(value) * interval.count
            result = append(result, fmt.Sprintf("%d%s", value, interval.name))
        }
    }
    if len(result) > 3 {
        result = result[:3]
    }
    return strings.Join(result, " ")
}

// statsHandler - handle http requests to /stats/<data_type> 
func statsHandler(w http.ResponseWriter, r *http.Request) {
    dataType := r.URL.Path[len("/stats/"):]
    registry := map[string]func() (interface{}, error){
        "du":        getDU,
        "temps":     getTemps,
        "boot_time": getBootTime,
    }
    if _, ok := registry[dataType];!ok {
        response := makeJSONResponse(522, "wrong data type")
        json.NewEncoder(w).Encode(response)
        return
    }
    data, err := registry[dataType]()
    if err != nil {
        response := makeJSONResponse(404, err.Error())
        json.NewEncoder(w).Encode(response)
        return
    }
    response := makeJSONResponse(0, data)
    json.NewEncoder(w).Encode(response)
}

func main() {
    // cmdline arguments
    hostPtr := flag.String("h", "127.0.0.1", "Server listen address")
    portPtr := flag.String("p", "9090", "Server listen port")
    flag.Parse()

    address := fmt.Sprintf("%s:%s", *hostPtr, *portPtr)

    http.HandleFunc("/stats/", statsHandler)
    log.Printf("Starting server on %s", address)
    log.Fatal(http.ListenAndServe(address, nil))
}

