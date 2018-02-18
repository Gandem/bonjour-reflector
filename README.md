# bonjour-reflector

## About this project

This project aims to make Bonjour devices (such as printers and chromecasts) discoverable and usable by other devices located in different VLANs.
This is similar to [avahi-reflector](http://www.avahi.org/), but in addition permits a fine-grained control of how Bonjour traffic is reflected between VLANs.

How this is done:
- A device searching for Bonjour devices sends mDNS packets.
- The bonjour-reflector intercepts these mDNS packets.
- bonjour-reflector gets the destination MAC address and changes the VLAN tag of the packet to the tag of the VLAN the corresponding Bonjour device is in.
- bonjour-reflector then sends the modified packet back on the network.
- The Bonjour device recieves the packet and sends a response which is also intercepted by bonjour-reflector.
- bonjour-reflector checks which VLANs are shared with the Bonjour device (identified by the source MAC address of the packet), and sends a packet to each of these VLANs.

For this to work, bonjour-reflector should have an interface with all the concerned VLANs on tagged.

The packet forwarding is limited to devices on VLANs whose access to a given bonjour device has been allowed. A configuration file lists, for each Bonjour device, which VLANs have access to it.

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