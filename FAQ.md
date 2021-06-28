# 1. How would you prove the code is correct?

With current time restraints, I have relied solely on unit testing as a way to validate that the inputs (i.e the packets) create the expected outputs (i.e the iptables rules are created). More specifically:

- The expected packets are parsed, and the unexpected packets are discarded
- The expected packets are inserted into the tracker
- The tracker correctly identifies source IPs that scanned more than 3 ports
- The firewall inserts the correct rules 

There's decent coverage size (the percentage looks small, but that's because it's a tiny program)

I have also tested this manually to verify the behaviour, using live traffic. 

# 2. How would you make this solution better?

- Statically link the pcap library, so users didn't need to supply their own. This was omitted to reduce scope creep. I would use a smaller base image for Docker.
- Not use iptables as the mechanism to block packets
  - I would consider using [nftables](https://github.com/google/nftables) 
  - or; using eBPF, and XDP actions like `XDP_DROP` to drop packets quickly. I didn't do this, because while I can recognise it's (probably) the superior option, especially for performance, I opted for simplicity.
-  Publish ready made docker containers to a registry (See Question 8)
-  Implement different backends, such as:
   -  a distributed datastore (etcd?) so that many probes can feed into a central store, where hosts could read from and block source IPs pre-emptively, before even seeing a packet.
   -  swappable firewalls (today iptables, tomorrow nftables, one day eBPF) and let the user select which one with flags
- Add notification integration (email, slack, etc)

# 3. Is it possible for this program to miss a connection?

Yes, absolutely. 

Firstly, it's possible libpcap could drop them packets because there was no room in the OS's buffer when they arrived, it essentially means the packets are not being read fast enough.

Once the packet is forwarded to the core parsing/tracking component of contrackr I am reasonably confident that contrackr won't drop them in typical cases. Note that contrackr will purposefully drop any connection that is not an inbound TCP-SYN packet (though this is unlikely as the unit tests confirm that BPF filter captures the right packets). The tracker, whilst just a hash map does have (untested) limitations in how fast it can add entries. It's possible that the unbuffered channels that are used will grow to a large size under a DOS, and that the server will go out of memory. This will crash the software, and it will of course miss a connection (because it won't be running)

# 4. If you weren't following these requirements, how would you solve the problem of logging every new connection?

I would use port mirroring (a concept that exists in the [Cloud too](https://aws.amazon.com/blogs/aws/new-vpc-traffic-mirroring/)) so that contrackr sniffs packets passively, buffers the logs to disk and sends them to a logging endpoint, such as syslog-ng, or logstash. Depending on the needs of the business, to ensure high availability and If I was using AWS, I would put the contrackr instances behind a loadbalancer (which is a valid target for VPC traffic mirroring) and ensure that I had appropriate health checking (such as a `/health` endpoint that returns HTTP 200)

Having the packet logging instance out of the serving path means that when it goes down it only affects packet logging, and not whatever else is running on the server along side contrackr.

# 5. Why did you choose x to write the build automation?

I chose Bazel for the following reasons:

**Incremental builds**

Bazel only rebuilds what is necessary, which helps ensure subsequent builds are fast. Furthermore, it supports distributed caching so when the CI/CD pipeline kicks in, it can reuse the cache from previous builds, and all build hosts can benefit. This same cache could be used for all developers of the project (ignoring that I'm the sole developer for this, and its a tiny project)

**Reproducability**

Bazel builds deterministic outputs. The output is the same every time, no matter how many times it's ran, no matter who runs it. This can not be said for Make files and Dockerfiles as they can run abritary commands and yield unpredictable results.

I consider Reproducible builds to be a tenet of Security.

**Upkeep is easy(ish)/Lots of Integrations**

Depending on the project there is a slight curve to setting up a repo, but ongoing maintainence is actually quite easy. For typical operations (adding new dependencies/libraries/code) this is automated via [Gazelle](https://github.com/bazelbuild/bazel-gazelle). Migrating from one CI/CD platform to another shouldn't be too hard either, since all you are running is `bazel build` and it returns appropriate exit statuses. There are many skylark macros for doing common operations (see an example for [containers](https://github.com/bazelbuild/rules_docker))

(I could go on, I tried to keep this brief)

# 6. Is there anything else you would test if you had more time?

Yes, I would:

- leverage fuzz test frameworks to generate semi-random pcaps to ensure that unexpected inputs do not break core components of contrackr (the parser, tracker and blocker in particular).
- use a mutation framework to find ineffective unit tests, such as [go-mutesting](https://github.com/zimmski/go-mutesting) (I have no experience with this particular library, but it seems OK at first glance)
- Benchmarks, to better understand the limitations of performance

# 7. What is the most important tool, script, or technique you have for solving problems in production? Explain why this tool/script/technique is the most important.

I'm a huge fan of `strace`. When in production I'm often faced with unfamiliar software, that in some cases, could be considered a blackbox. Without documentation, the source code, or even appropriate logging. I need to better understand _what_ the application is actually doing.

To do this, I can run `strace` and inspect the system calls, which help me understand how the program is interacting with the kernel. It can help me to answer questions like:

- Is the software reading, and writing files?
  - if so, where? What read/write flags? Is it getting a permission error?
  - Particularly useful if you know it's writing log files, but not sure where (yes, lsof exists)
  - what are the contents of the reads/writes?
- Is the software reading, or writing to network sockets?
  - what are their destinations? what ports? what responses?
- Is the program calling other programs (`exec`), which ones?

It of course doesn't replace the myriad of other tools and commands for troubleshooting in production, but you can get surprisingly far with just `strace`.


# 8. If you had to deploy this program to hundreds of servers, what would be your preferred method? Why?

I've currently got some basic [release automation](https://github.com/michaelmcallister/contrackr/blob/main/.github/workflows/release.yml) set up that when a tagged release is created the build artefacts (the binary, and container filesystem) are added to the releases page on GitHub. 

I would extend this automation to:
- push the image to a container registry
- generate a deployment manifest (likely a DaemonSet), and 
- deploy (see an example [GitHub](https://github.com/marketplace/actions/deploy-to-kubernetes-cluster) action)
  
Another reason for using Bazel is that there are a lot of Skylark macros for common operations. For instance, pushing to a container registry is as simple as creating a [`container_push`](https://github.com/bazelbuild/rules_docker#container_push) target.

This would be my preferred method as it would be easy to setup, easy to reason about (the configs are small), and mostly automated (push on green). In order to get to this, the tests would need to be written better.

# 9.  What is the hardest technical problem or outage you've had to solve in your career? Explain what made it so difficult?

The most challenging outage I've had to solve was when I was working at an ISP (~1 million subscribers) and due to the heat during the day (36°C/96.8°F) one of our datacentres AC system started to fail. This datacentre was one of 2 main sites in the state, and hosted services such as mail, website hosting, and RADIUS authentication servers, amongst other internal ones.

My job was to figure out how to safely power down racks and networking equipment to reduce the ambient temperature and protect the equipment, whilst not causing too much impact (i.e turning off non-essential servers). I also had to spin up capacity, and migrate servers to our other datacentre.

In addition to this, I had to constantly relay updates to execs and our comms team much of it persisted [in this thread](https://forums.whirlpool.net.au/archive/2355540)

Even our own internally hosted status page went down, and I had to very quickly put something together and host it on AWS S3, so that users could see updates quickly. Most of the state was offline, for about 4 hours.

It was difficult for a combination of:
- relatively large impact
- poor DR strategy meant there was no "playbook" to run. I just had to figure out which racks to power down, what capacity needs to be spun up, etc.
- I was relatively new to the industry and inexperienced
- I had to juggle fixing problems and providing meaningful updates to execs at the same time.