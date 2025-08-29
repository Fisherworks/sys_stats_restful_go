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
    "github.com/shirou/gopsutil/v4/host"
    "github.com/shirou/gopsutil/v4/load"
    "github.com/shirou/gopsutil/v4/mem"
    "github.com/shirou/gopsutil/v4/sensors"
)

// APIResponse defines the structure of the API response
type APIResponse struct {
    Code    int         `json:"code"`
    Status  string      `json:"status"`
    Data    interface{} `json:"data"`
}

// Define the API status code enumeration
var API_MSG_ENUM = map[int]string{
    0:    "success",
    404:  "no data",
    522:  "wrong arguments",
}

// makeJSONResponse is used to format the JSON response
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

// roundToNDecimal is a helper function to round a floating-point number to n decimal places
func roundToNDecimal(num float64, n int) float64 {
    multiplier := math.Pow(10, float64(n))
    return math.Round(num*multiplier) / multiplier
}

// getDU retrieves the disk usage information for all mount points
func getDU() (interface{}, error) {
    partitions, err := disk.Partitions(true)
    if err != nil {
        log.Printf("Error getting disk partitions: %v", err)
        return nil, err
    }

    result := make(map[string]map[string]float64)
    for _, partition := range partitions {
        device := partition.Device
        if strings.HasPrefix(device, "/dev/") || strings.Contains(device, ":") {
            usage, err := disk.Usage(partition.Mountpoint)
            if err != nil {
                log.Printf("Error getting disk usage for %s: %v", partition.Mountpoint, err)
                continue
            }
            result[partition.Mountpoint] = map[string]float64{
                "free_size": roundToNDecimal(float64(usage.Free)/(1024*1024), 2),
                "free_rate": roundToNDecimal(100-float64(usage.UsedPercent), 2),
            }
        }
    }
    return result, nil
}

// getTemps retrieves the sensor temperature information with warning details
func getTemps() (interface{}, error) {
    temps, err := sensors.SensorsTemperatures()

    // Handle warnings from gopsutil's Warnings type (pointer receiver)
    if err != nil {
        log.Printf("Error getting sensors temperatures: %v", err)
        if warns, ok := err.(*sensors.Warnings); ok {
            for i, warn := range warns.List {
                log.Printf("Sensor warning %d: %v", i+1, warn)
            }
        } else {
            log.Printf("Sensor error: %v", err)
        }
    }

    // Return error if no temperature data available
    if len(temps) == 0 {
        return nil, fmt.Errorf("no valid temperature data: %w", err)
    }

    // Process valid temperature data
    result := make(map[string]map[string]float64)
    for _, temp := range temps {
        result[temp.SensorKey] = map[string]float64{
            "curr": roundToNDecimal(temp.Temperature, 2),
            "crit": roundToNDecimal(temp.Critical, 2),
        }
    }
    return result, nil
}

// getBootTime retrieves the system boot time information
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

// formatElapsedTime converts the number of seconds into a human-readable format
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

// getSysLoads retrieves the 1-minute and 15-minute average system load information
func getSysLoads() (interface{}, error) {
    loadInfo, err := load.Avg()
    if err != nil {
        log.Printf("Error getting system load: %v", err)
        return nil, err
    }
    return map[string]float64{
        "load_01": roundToNDecimal(loadInfo.Load1, 2),
        "load_05": roundToNDecimal(loadInfo.Load5, 2),
        "load_15": roundToNDecimal(loadInfo.Load15, 2),
    }, nil
}

// getMemInfo retrieves the OS virtual memory information
func getMemInfo() (interface{}, error) {
    memInfo, err := mem.VirtualMemory()
    if err != nil {
        log.Printf("Error getting mem info: %v", err)
        return nil, err
    }
    return map[string]float64{
        "total":     roundToNDecimal(float64(memInfo.Total)/(1024*1024), 2),
        "used":      roundToNDecimal(float64(memInfo.Used)/(1024*1024), 2),
        "free_rate": roundToNDecimal(100-float64(memInfo.UsedPercent), 2),
    }, nil
}

// statsHandler handles the /stats/<data_type> route
func statsHandler(w http.ResponseWriter, r *http.Request) {
    dataType := r.URL.Path[len("/stats/"):]
    registry := map[string]func() (interface{}, error){
        "disk_usage": getDU,
        "sensors_temp": getTemps,
        "boot_time": getBootTime,
        "load_avg":  getSysLoads,
        "mem_info":  getMemInfo,
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
    // Define command-line arguments
    hostPtr := flag.String("h", "127.0.0.1", "Server listen address")
    portPtr := flag.String("p", "9090", "Server listen port")
    flag.Parse()

    address := fmt.Sprintf("%s:%s", *hostPtr, *portPtr)

    http.HandleFunc("/stats/", statsHandler)
    log.Printf("Starting server on %s", address)
    log.Fatal(http.ListenAndServe(address, nil))
}

