# bonjour-reflector

## About this project

This projects aims to make bonjour devices (such as printers, chromecasts, ...) discoverable and usable by other devices located in different VLANs.

This is done by intercepting mDNS packets and forwarding them to the correct VLANs (for the request packets, the VLANs where the bonjour devices are located; for the response packets, the VLANs where the searching devices are located).

The packet forwarding is limited to devices on VLANs whose access to a given bonjour device has been allowed. A configuration file lists all shared devices with the VLANs each of them are shared with.

## Installation

You need [dep](https://github.com/golang/dep) to install the project dependencies.
Once you've installed dep, run:

```
dep ensure
```

to install the dependencies.

One of the dependencies of the project (gopacket/pcap) also needs the libpcap header files to work properly.
On Linux-based distributions, you can do this by installing the development version of libpcap.


## App setup

First, indicate in the `config.toml` file which of your network interfaces you want to listen to.

Then build the package:

```
go build
```

And run

```
./bonjour-reflector -config=./config.toml
```

(you may need to run this line with administrator privileges to listen to your interface).

You may use any configuration file you want (following the same structure as the template `./config.toml` file provided) by specifying its path with the `-config` option.
