# bonjour-reflector

## Installation

You need [dep](https://github.com/golang/dep) to install the project dependancies.
Once you've installed dep, run:

```
dep ensure
```

to install the dependencies.

One of the dependencies of the project (gopacket/pcap) also needs the libpcap header files to work properly.
On Linux-based distributions, you can do this by installing the development version of libpcap.


## Setup

First, indicate in the `config.toml` file which of your network interfaces you want to listen to.

Then build the package:

```
go build
```

And run `./bonjour-reflector` (you may need to run it with administrator privileges to listen to your interface).