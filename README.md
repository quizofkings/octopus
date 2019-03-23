# octopus
[![Go Report Card](https://goreportcard.com/badge/github.com/quizofkings/octopus)](https://goreportcard.com/report/github.com/quizofkings/octopus)
[![CircleCI](https://circleci.com/gh/quizofkings/octopus/tree/master.svg?style=svg)](https://circleci.com/gh/quizofkings/octopus/tree/master)

Redis Multi-Cluster load balancing using prefix key,
No need to change anything, octopus use RESP (REdis Serialization Protocol).
*Developing/ Testing/ Documention in progress*

## Config
* etcd
* consul
* json file

Sample JSON:
```json
{
    "port": 9090,
    "pool": {
        "maxcap": 10,
        "initcap": 3
    },
    "clusters": [
        {
            "ismain": true,
            "name": "qok-match",
            "nodes": [
                "172.20.3.4:7001",
                "172.20.3.10:7004"
            ],
            "prefix": [
                "match:",
                "stats:"
            ]
        },
        {
            "ismain": false,
            "name": "qok-tournament",
            "nodes": [
                "10.5.150.5:7000",
                "10.5.150.12:7003"
            ],
            "prefix": [
                "tournament:"
            ]
        }
    ]
}
```


## Usage
```sh
Flags:
      --consulkey string   consul k/v namespace
      --etcdaddr string    etcd file address
  -f, --file string        config file name (without extenstion) (default "config")
  -h, --help               help for gate
      --host string        config remote host (ex. http://127.0.0.1:4001)
  -p, --path string        config path (default is /) (default "/")
  -r, --remotekey string   etcd/consul remote key (default is empty)
  -t, --type string        config type (file, etcd, consul) (default "file")
```