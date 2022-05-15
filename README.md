# vwap

This project calculate the volume-weighted average price of various trading crypto pairs in real time. For now its uses
the coinbase websocket for the live transaction data. For now this project only calculates the vwap for BTC-USD, ETH-USD 
and ETH-BTC trading pairs for 200 data pairs at max.

<h3>How To Run</h3>

- To run the application with default configuration, run `make run`
- To run the application in debug mode, run `make debug`

<h3>Architecture</h3>

![](../Desktop/vwap arch.png)

There are mainly 3 parts of the system executing concurrently:

1. Writer:

   - Writer writes the initialization and termination message to the coinbase websocket. It consumes the message from
    output queue and send it to coinbase. 
   
2. Reader:

   - Reader reads the data from the coinbase websocket continuously, parses it, and pushes it to the vwap calculation
   engine where it calculates the vwap of the consumed pair

3. Driver Program:

    - Driver program spawns and monitors the overall execution of readers, writers, establishing connection etc.

VWAP Calculation Engine:

This is the heart of the project where all the calculation happens. It keeps track of the volumes and weighted price of
all trades. We are using the circular buffer to efficiently use the memory buffer and replace the existing memory when we
reach the overflow. It reduces the memory footprint of algorithm. The time complexity of the algoirthm is contant, 
i.e O(1) and memory complexity is O(capacity), i.e constant 200 for the given program.

<h3>External Dependencies: </h3>

- This project uses the coinbase websocket dependency for the live trade data. It subscribes to the matches channel.
https://docs.cloud.coinbase.com/exchange/docs/websocket-channels#matches-channel
