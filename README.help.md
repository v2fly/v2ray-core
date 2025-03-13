# 配置


## 透明代理
https://toutyrater.github.io/app/tproxy.html

https://www.v2fly.org/config/dns.html#dnsobject
## 代理转发
https://guide.v2fly.org/advanced/outboundproxy.html

## API添删用户

## 配置合并
```
./elink convert  samples/in_*.json  samples/out_*.json samples/log.json samples/route.json samples/dns.json 
```

## 测试代理
```
curl -vv --socks5-hostname test:password@127.0.0.1:34560 https://ifconfig.me
```




