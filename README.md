A Simple Go Sub-Domain Based Router
==
My initial goal is to create a very basic domain-based routing. It doesn't take up much memory, and can handle forwarding all requests without
# Setup
Have GO version 1.7.4 installed then run
```
go get ./...
```

Create config.json
```
cp config.json.sample config.json
```

Edit config.json to represent your subdomains and point to appropriate machine IP addresses

Note: the regex allows specific url requests, if the url request does not match the regex it is rejected