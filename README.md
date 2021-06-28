<h1 align="center">michaelmcallister/contrackr</h1>

<p align="center">
    <a href="https://github.com/michaelmcallister/contrackr/actions?query=workflow%3Abuild">
        <img alt="Build Status" src="https://github.com/michaelmcallister/contrackr/workflows/build/badge.svg">
    </a>
      <a href="https://codecov.io/gh/michaelmcallister/contrackr">
        <img src="https://codecov.io/gh/michaelmcallister/contrackr/branch/main/graph/badge.svg?token=S0V4HRd7Bo"/>
      </a>
    <a href="https://goreportcard.com/report/michaelmcallister/contrackr">
        <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/michaelmcallister/contrackr">
    </a>
</p>


## Intro
This is a simple server that listens for incoming TCP connections and blocks source IPs that attempt to connect to 3 different ports within a minute. It uses iptables as the firewall.

## Dependencies

To run contrackr you will need both `iptables` and `libpcap`, to install them
run 

Ubuntu:
```
$ apt-get install iptables libpcap
```

Arch:
```
$ pacman -Syu iptables libpcap
```

## Installing

If you'd like to install the latest release you have two options:

**Prebuilt Linux binary**

You can find the binary for the latest release below:
- https://github.com/michaelmcallister/contrackr/releases/download/latest/contrackr

Please note that this binary is not statically linked, you will still need libpcap.

**Docker Image**

Run the following command to import the Docker filesystem image:

```
$ wget https://github.com/michaelmcallister/contrackr/releases/download/latest/contrackr_image.tar
$ docker load -i contrackr_image.tar    
```

## Building

You will need libpcap header files to build with either Bazel, or Go to install them

Ubuntu:
```
$ apt-get install libpcap-dev
```

Arch:
```
$ pacman -Syu libpcap-dev
```

## With Bazel (optional if not building Docker image)

**Installing Bazel**

If you do not have Bazel installed, the easiest way to install is via
[Bazelisk](https://github.com/bazelbuild/bazelisk), which is a user-friendly
launcher for Bazel. To install Bazelisk you have a few options:

1) `npm install -g @bazel/bazelisk`
2) `go get github.com/bazelbuild/bazelisk`
3) using Homebrew on macOS
4) using a [binary release](https://github.com/bazelbuild/bazelisk/releases) for Linux, macOS, or Windows

It's also in the AUR (I use Arch btw)

**Building Binary**

Run bazel build for the binary target
```
$ bazel build //cmd:contrackr 
```
The resulting binary will be available in the below path under `bazel-bin/`
```
$ bazel-bin/cmd/contrackr_/contrackr 
```

## Using Go
Simply run 

`$ go build cmd/contrackr.go` 

the binary will be in the current working directory as `contrackr`

## Building Docker Container

Bazel will build a Docker image tarball that is suitable for importing via 
`docker import`

To build the tarball and import it as an image on the host Docker run the following  (note this is `bazel run`, not `bazel build`). The current user must have permissions to connect to the hosts Docker daemon.

```
$ bazel run :contrackr_image
```

You will then find the resulting image in Docker under the `bazel` repository and the `contrackr_image` tag:

```
$ sudo docker images bazel:contrackr_image
REPOSITORY   TAG               IMAGE ID       CREATED        SIZE
bazel        contrackr_image   df280f4d4b2e   51 years ago   96.8MB   
```

If you'd prefer to just build the tarball and import it yourself, run 

```
$ bazel build :contrackr_image 
```
The resulting image tar will be available in the below path under `bazel-bin/`
```
$ bazel-bin/contrackr_image-layer.tar
```

## Running

### Binary

To run, simply supply the interface (eg. eth0) you'd like to capture packets on with the `-i` flag.
You can also supply `-i any` to listen on any interface.

For logging add the `-logtostderr=true` flag, and if need be increase the verbosity with `-v 2`

*Running as non-root*

As contrackr uses iptables to manipulate the host firewall it requires root. There are possible workarounds as [documented here](https://dbpilot.net/2018/3-ways-to-run-iptables-l-as-non-root-user/)

Once you have worked around iptables, you can set the capabilities on the binary itself, like so
```
$ setcap cap_net_admin,cap_net_raw+ep ./contrackr 
```
and you will be able to run contrackr without issue.

### Docker

To manipulate the host firewall and capture the packets appropriately you will
need to pass in `--net=host --cap-add=NET_ADMIN` to your Docker run command. Like so
```
$ docker run --net=host --cap-add=NET_ADMIN -t bazel:contrackr_image -i wlp3s0
```

This will capture packets from the host, and manipulate iptables as appropriate.

### Contributing

Whilst it looks intimidating, the Bazel build rules are mostly managed by [Gazelle](https://github.com/bazelbuild/bazel-gazelle). It will take care of updating the `BUILD.bazel` files for you. You do not need to install any dependencies other than Bazel.
 
If you change any dependencies within Go, simply run:

```
$ bazel run //:gazelle -- update 
```

When you submit a PR the [CI/CD pipeline](https://github.com/michaelmcallister/contrackr/actions/workflows/ci.yml) will ensure it builds and tests correctly.

## Common Issues

If you run into an error like below:

```
can't initialize ip6tables table `filter': Table does not exist (do you need to insmod?)
```

Make sure you run `sudo modprobe ip6table_filter`

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details