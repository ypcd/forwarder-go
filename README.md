# forwarder-go
A high-performance network forwarder developed in go language.


forwarder-go is a high-performance network forwarder developed in go language.

Support tcp protocol.

Adopt the csp concurrency model, the net library in the go language standard library.

The coroutine pool technology is adopted to greatly reduce the number of coroutines and memory usage.


Supported operating systems: windows, linux, macos


## Performance:

### qps: more than 40000/s

(10 million connections, single connection data volume 4Kbytes)

### Throughput: more than 1000Mbytes/s

(Single connection data volume 1Gbytes)

illustrate:

To further improve the network performance of forwarder-go, it is necessary to use lower-level network technologies such as zero copy or DPDK.

In the short term, forwarder-go will not adopt such underlying network technology.

If you want to adopt the above-mentioned underlying network technology, you will use c/c++ or rust as the development language.


### test environment:

ubuntu 22.04.2

Intel i5 1135g7 2.4GHz

4 cores 4 threads without hyperthreading


## New features that will be added:

support udp protocol

Support load balancing

access control (ip)

network speed limit
