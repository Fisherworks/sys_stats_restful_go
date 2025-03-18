# Sys Stats Restful - Go

A Go based porting repo of my other one - [Sys Stats Restful](https://github.com/Fisherworks/sys_stats_restful). And this is basically done with the support by doubao AI agent, which gives me a brief intro of Go Lang and shows me the overall differences against Python. 

## Why Go

Well, maybe I just want to try something new that can be both easily written/read and deployed anywhere, such as an embedded device like a router, without concerning too much about its dependencies. Furthermore, considering the barely simplified features of this app, it might be a nice route to a new learning experience. 

## More than python version
New `data_type` keywords added, it's 5 for them now, including `mem, load_avg, du, boot_time, temps`. 
```
$ curl http://localhost:9090/stats/mem | jq .
{
  "code": 0,
  "status": "success",
  "data": {
    "free_rate": 59.94,
    "total": "1962.04M",
    "used": "786.02M"
  }
}
```

## Architectures Tested

* Linux - aarch64/arm64
* Linux - amd64
