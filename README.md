# zyzzyva

[Zyzzyva: Speculative Byzantine Fault Tolerance](http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.122.112&rep=rep1&type=pdf) impl

## Deploy

Example:

```bash
docker run -d --name zyzzyva -e IP_PREFIX=10.0.0.1 --network host myl7/zyzzyva -id 0
```

Requires N = 3 * F + 1 servers and M clients, server or client is decided by whether id >= N

## License

MIT
