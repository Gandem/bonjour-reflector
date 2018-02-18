# bonjour-reflector

## About this project

Bonjour-reflector makes Bonjour devices such as printers, Chromecasts or Spotify Connect speakers, discoverable and usable by other devices located on different VLANs.

Compared to other tools such as [avahi-reflector](http://www.avahi.org/), Bonjour-reflector allows a more fine-grained control of how Bonjour traffic is reflected across VLANs. 

## How it works

Bonjour-reflector works by intercepting all mDNS traffic and rewriting layers 2 and 3 of the packets to reflect them across the appropriate VLANs.

A configuration file lists, for each Bonjour device (defined by its MAC address), which VLANs should have access to this device. mDNS packets will only be forwarded if the configuration file says so.

The interface on which Bonjour-reflector runs should be configured so that it receives each VLAN's traffic, tagged.

In detail, here is what happens when Bonjour-reflector runs:
- a device searching for Bonjour devices sends mDNS packets on his VLAN.
- bonjour-reflector receives these mDNS packets, tagged with the original VLAN.
- bonjour-reflector looks up in its configuration to which VLANs it should forward the mDNS request, and send new packets tagged with these new VLANs.
- one Bonjour device receives the packet and sends a response which is also intercepted by bonjour-reflector.
- bonjour-reflector reads the source MAC of the mDNS response, looks up in its configuration which VLANs are shared with this Bonjour device, and reflects the mDNS response on each of these VLANs.

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

## Contribution

Help on this project is very welcomed. Before submitting your contribution, please make sure to take a moment and read through the following guidelines:

- The `master` branch contains the latest stable version of the project. All development should be done in dedicated branches.
- Try to name your branch in a clear way, for example by following this pattern: `username/what-i-am-fixing`.
- Do not check in any compiled binaries in the commits.
- It's okay to have multiple small commits as you work on the PR - we will squash them before merging.
- Make sure all test cases pass (using `go test`).
- When fixing a bug:
    - Prefix your PR with `Fix:`, and add references to the issues linked to your PR (if they exist),
    - Add test coverage if applicable.
- When adding a new feature:
    - Prefix your PR with `Feature:`,
    - Add a description of your feature and reasons to add this feature,
    - Add test cases for this feature.


## License

MIT
