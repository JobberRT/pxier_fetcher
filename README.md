# pxier_fetcher
`pxier_fetcher` is the proxy fetcher for [Pxier](https://github.com/JobberRT/pxier)

## Configuration
`factory.fetch_interval`: fetch proxy interval
`factory.selected_executor`: which kinds of provider you choose. You can choose these(case not sensitive):
- cpl (From: https://github.com/clarketm/proxy-list)
- tsx (From: https://github.com/TheSpeedX/PROXY-List)
- str (From: https://github.com/ShiftyTR/Proxy-List)
- ihuan (From: https://ip.ihuan.me/ti.html)

`executor.XXX.timeout`: proxy fetch timeout
`executor.XXX.each_fetch_num`: how many proxies for each fetch request
`executor.XXX.proxy`: set proxy for fetching proxy

Other settings don't need to be changed.

## How to use
Recommend to use [Pxier](https://github.com/JobberRT/pxier) README's docker-compose file to deploy. Otherwise, you can compile and change the configuration and rename the `config.example.yaml` to `config.yaml`, then you can start the executable.