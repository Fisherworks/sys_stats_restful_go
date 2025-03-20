# Sys Stats Restful - Go

A Go based porting repo of my other one - [Sys Stats Restful](https://github.com/Fisherworks/sys_stats_restful). And this is basically done with the support by doubao AI agent, which gives me a brief intro of Go Lang and shows me the overall differences against Python. 

## Why Go

Well, maybe I just want to try something new that can be both easily written/read and deployed anywhere, such as an embedded device like a router, without concerning too much about its dependencies. Furthermore, considering the barely simplified features of this app, it might be a nice route to a new learning experience. 

## More than Python version
New `data_type` keywords added, it's 5 of them now, including `mem_info, load_avg, disk_usage, boot_time, sensors_temp`. 
```
$ curl http://localhost:9090/stats/disk_usage | jq .
{
  "code": 0,
  "status": "success",
  "data": {
    "/": {
      "free_rate": 86.18,
      "free_size": 64220.65
    },
    "/boot": {
      "free_rate": 52.63,
      "free_size": 268.91
    }
  }
}
```

## How to use
Download and run, good to go.
```
$ ./entry -h 0.0.0.0 -p 9090
2025/03/18 12:34:56 Starting server on 0.0.0.0:9090
```

## Architectures Tested

* Linux - aarch64/arm64
* Linux - amd64
* Win10 - amd64 (Partial)
