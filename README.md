# inetb

inetb is a benchmarking tool for BGP RIB / FIB convergence

## Demo

TBU


## Overview

inetb assumes a test environment below:

TODO: put image

* router to be tested(DUT) has 2 eBGP sessions to gobgpds
  * gobgpds are configured with different ASNs
  * gobgpds run on the same server

inetb counts BGP updates:

* inetb controls gobgpds via gRPC
  * Make gobgpds send and receive many BGP updates to and from the router(DUT)
* inetb monitors which prefixes are advertised or withdrawn on peering interfaces with libpcap, without interrupting gobgpd behaviors

Then inetb dumps the results with timestamp to JSON files.

You can draw charts with your favorite libraries. An example of convergence time charts is included in this repository.


### Scenario 1: The router(DUT) receives new routes

1. gobgpd1 advertises new routes to the router(DUT)
  * inetb counts NLRIs in BGP updates passed through router(DUT) - gobgpd1 peer
2. After router(DUT) gets converged for the new NLRI, it should advertise the same NLRIs to gobgpd2
  * inetb also counts NLRIs on router(DUT) - gobgpd2 peer

### Scenario 2: The router(DUT) receives stronger routes

1. gobgpd2 advertises stronger routes to the router(DUT)
  * inetb counts NLRIs in BGP updates passed through router(DUT) - gobgpd2 peer
2. After router(DUT) gets converged for the new NLRI, it should advertise the same NLRIs to gobgpd1
  * inetb also counts NLRIs on router(DUT) - gobgpd1 peer

### Scenario 3: The router(DUT) receives withdrawals of stronger routes

1. gobgpd2 sends withdrawals of stronger routes to the router(DUT)
  * inetb counts withdrawn prefixes in BGP updates passed through router(DUT) - gobgpd2 peer
2. After router(DUT) gets converged for the withdrawn BGP updates, it should send the same withdrawals back to gobgpd1
  * inetb also counts withdrawals on router(DUT) - gobgpd1 peer

### Scenario 4: The router(DUT) receives withdrawals of the rest

1. gobgpd1 sends withdrawals of all routes to the router(DUT)
  * inetb counts withdrawn prefixes in BGP updates passed through router(DUT) - gobgpd1 peer
2. After router(DUT) gets converged for the withdrawn BGP updates, it should send the same withdrawals back to gobgpd2
  * inetb also counts withdrawals on router(DUT) - gobgpd2 peer


## Installation

```
go get github.com/codeout/inetb
```


## How to Use

1 . Start 2 gobgpds with different ASNs.
  * See [gobgp document](https://github.com/osrg/gobgp/) for details
  * Config example is available [here](gobgpd.conf)

:warning: NOTE:

* gobgpds should be configured with local-address so that inetb can identify the 2 BGP sessions

```
gobgpd -f gobgpd1.conf --disable-stdlog
gobgpd -f gobgpd2.conf --disable-stdlog --api-hosts :50052
```

2 . Download MRT File

Table dump v2 for full routes is available at [Route Views Project](http://archive.routeviews.org/) for instance.

3 . Start inetb

```
inetb rib.20170707.1200
```

It will examine 4 test scenarios above and then dump JSON reports in ```reports/``` directory.


## Features to be implemented

* FIB convergence benchmark


## Copyright and License

Copyright (c) 2017 Shintaro Kojima. Code released under the [MIT license](LICENSE.txt).

Code includes a part of [gobgp](https://github.com/osrg/gobgp/) which is distributed in the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
