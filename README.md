# student_24_mino_libp2p

```
$ go get demo
$ go build
```

```
$ ./demo -create 
2024/03/19 16:00:56 creating host...
2024/03/19 16:00:56 created host 12D3KooWGx7Dnsp1PBh46YWDji6SP25tooPRebVXgKDLj4y5atAT
2024/03/19 16:00:56 storing identity to file...
2024/03/19 16:00:56 stored identity to file 12D3KooWGx7Dnsp1PBh46YWDji6SP25tooPRebVXgKDLj4y5atAT.txt
Your address is: [/ip4/127.0.0.1/tcp/59214/ws /ip4/128.179.211.43/tcp/59214/ws]
Run -connect "12D3KooWGx7Dnsp1PBh46YWDji6SP25tooPRebVXgKDLj4y5atAT[/ip4/127.0.0.1/tcp/59214/ws /ip4/128.179.211.43/tcp/59214/ws]" "destination1" "destination2" ...

```

```
$ ./demo -connect "12D3KooWGx7Dnsp1PBh46YWDji6SP25tooPRebVXgKDLj4y5atAT[/ip4/127.0.0.1/tcp/59214/ws /ip4/128.179.211.43/tcp/59214/ws]" "12D3KooWCDUWPijNkwMskv5Dyap2SggvEVoB8JTCmFyaWotnZvt8[/ip4/127.0.0.1/tcp/59217/ws /ip4/128.179.211.43/tcp/59217/ws]"
2024/03/19 16:04:37 starting host 12D3KooWGx7Dnsp1PBh46YWDji6SP25tooPRebVXgKDLj4y5atAT at [/ip4/127.0.0.1/tcp/59214/ws /ip4/128.179.211.43/tcp/59214/ws]...
2024/03/19 16:04:37 started host 12D3KooWGx7Dnsp1PBh46YWDji6SP25tooPRebVXgKDLj4y5atAT at [/ip4/127.0.0.1/tcp/59214/ws /ip4/128.179.211.43/tcp/59214/ws]
2024/03/19 16:04:37 adding peers to peer store...
2024/03/19 16:04:37 added peers
2024/03/19 16:04:37 listening for streams...
2024/03/19 16:04:37 opening streams...
2024/03/19 16:04:37 could not open stream to 12D3KooWCDUWPijNkwMskv5Dyap2SggvEVoB8JTCmFyaWotnZvt8: failed to dial: failed to dial 12D3KooWCDUWPijNkwMskv5Dyap2SggvEVoB8JTCmFyaWotnZvt8: all dials failed
  * [/ip4/128.179.211.43/tcp/59217/ws] dial tcp 128.179.211.43:59217: connect: connection refused
  * [/ip4/127.0.0.1/tcp/59217/ws] dial tcp 127.0.0.1:59217: connect: connection refused
2024/03/19 16:05:02 sending/receiving ping...
sent ping: ping <peer.ID 12*y5atAT><peer.ID 12*tnZvt8>#1
received ping: ping <peer.ID 12*tnZvt8><peer.ID 12*y5atAT>#1

```