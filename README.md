# simple_chat

A simple websocket chat accomplished in self-education purpose.

Uses redis pub sub for messaging and can be scaled horizontally.

### usage

Build executable:
```shell script
make docker-build
```

Run locally on port 8080
```shell script
docker-compose up simple_chat
```

### tests

launch redis-tst:
```shell script
docker-compose up -d redis-tst
```

run tests:
```shell script
make test
```

### disclaimer

service is developed only for self-education purpose and exposes the number of simplifications and security issues.
