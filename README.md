
## WARNING!!!
This is NOT production-ready bullet-proof software, and is full of 
security vulnerabilities (timing channels, hard-coded secret keys, etc.).
that make it totally unsuitable for use in a production environment. 

## Background

This is the software prototype that accompanies the research paper:

> ["Prio: Private, Robust, and Scalable Computation of Aggregate Statistics"](https://crypto.stanford.edu/prio/paper.pdf)<br>
> by Henry Corrigan-Gibbs and Dan Boneh<br>
> NSDI 2017

For more information, please visit:
  [https://crypto.stanford.edu/prio/](https://crypto.stanford.edu/prio/).

Dependencies (for fast polynomial operations):
* [FLINT](http://www.flintlib.org/) 2.5.2
* [GMP](https://gmplib.org/) 6.1.1
* [MPFR](http://www.mpfr.org/) 3.1.4


# Getting started

## 1. Install dependencies

Make sure that you have Go installed.
We have tested Prio with Go 1.9.3.
Check that you have go installed correctly by running:

     $ go version

You should see something like:

     go version go1.9.3 darwin/amd64

Download and install the FLINT, GMP, and MPFR libraries (linked
above), and the git tools (needed for running "go get").

On Ubuntu, you should be able to install everything you need with

    # apt-get install golang-go git libgmp-dev libflint-dev libmpfr-dev

It's convenient to set your GOPATH to ~/go. To do so, run

    $ mkdir ~/go

and add the following to your .bashrc (or equivalent):

    $ export GOPATH=~/go

## 2. Install Prio

Download and install the Prio source:

    go get github.com/henrycg/prio/tclient
    go get github.com/henrycg/prio/tserver
    go get github.com/henrycg/prio/runservers

If everything works, you should now have binaries
tclient, tserver, and runservers in $GOPATH/bin.

## 3. Run the software

There are three relevant binaries:
* tclient - the Prio client 
* tserver - the Prio server 
* runservers - a utility program that starts up a cluster of Prio servers

Once you have compiled the Prio binaries, you should be able to run
a cluster of servers using the default configuration. To do so, run: 

    # cd to wherever your go binaries are.
    # On my machine, they are in $GOPATH/bin
    $ cd $GOPATH/bin 
    $ ./runservers

You should see a whole bunch of debugging output:

    2018/01/28 15:02:51 Working over field of 16 bits
    2018/01/28 15:02:51 Loading roots...
    2018/01/28 15:02:52 Working over field of 16 bits
    2018/01/28 15:02:52 Loading roots...
    2018/01/28 15:02:52 Working over field of 16 bits
    2018/01/28 15:02:52 Loading roots...
    2018/01/28 15:02:52 Working over field of 16 bits
    2018/01/28 15:02:52 Loading roots...
    2018/01/28 15:02:52 Working over field of 16 bits
    2018/01/28 15:02:52 Loading roots...
    2018/01/28 15:02:52 Working over field of 16 bits
    2018/01/28 15:02:52 Loading roots...
    [Server 1] 2018/01/28 15:02:53 Public RPC 1 is listening at 0.0.0.0:9001
    [Server 1] 2018/01/28 15:02:53 Private RPC 1 is listening at 0.0.0.0:9051
    [Server 1] 2018/01/28 15:02:53 Listening at 0.0.0.0:9001
    [Server 2] 2018/01/28 15:02:53 Public RPC 2 is listening at 0.0.0.0:9002
    [Server 2] 2018/01/28 15:02:53 Private RPC 2 is listening at 0.0.0.0:9052
    [Server 2] 2018/01/28 15:02:53 Listening at 0.0.0.0:9002
    [Server 4] 2018/01/28 15:02:53 Public RPC 4 is listening at 0.0.0.0:9004
    [Server 4] 2018/01/28 15:02:53 Private RPC 4 is listening at 0.0.0.0:9054
    [Server 4] 2018/01/28 15:02:53 Listening at 0.0.0.0:9004
    [Server 0] 2018/01/28 15:02:53 Public RPC 0 is listening at 0.0.0.0:9000
    [Server 0] 2018/01/28 15:02:53 Private RPC 0 is listening at 0.0.0.0:9050
    [Server 0] 2018/01/28 15:02:53 Listening at 0.0.0.0:9000
    [Server 3] 2018/01/28 15:02:53 Public RPC 3 is listening at 0.0.0.0:9003
    [Server 3] 2018/01/28 15:02:53 Private RPC 3 is listening at 0.0.0.0:9053
    [Server 3] 2018/01/28 15:02:53 Listening at 0.0.0.0:9003
    [Server 2] 2018/01/28 15:02:56 Certs [0xc4248be000]
    ...

If you get an error message of the form:

    [Server 1] 2018/01/28 15:09:31 Listener error:listen tcp 0.0.0.0:9001: bind: address already in use

Make sure that you do not have a "runservers" process already running.
To do so, run:

    $ killall tserver

which will kill all running go server processes.

Now you have a set of five Prio servers up and running and waiting for 
data submissions from clients.
Now we can run a Prio client to submit a request to the servers.
In a separate terminal window run:

    $ cd $GOPATH/bin
    $ ./tclient

You should see a bunch of output ending with something like:

    2018/01/28 15:08:42 Done generating args
    2018/01/28 15:08:43 Processed request 0

In the terminal window where the servers are running, you should
see a bunch of debugging messages scrolling by.
If you do, congratulations! You have a working Prio deployment.

## 4. More advanced usage

Each of the three Prio binaries takes a set of command-line flags.
To get a listing of the possible flags, run (for example):

    ./tclient --help 

or

    ./tserver --help

The most important argument to both the client and server binaries
is the "-config" flag, which tells the program where to find the 
configuration file for the Prio deployment.
The config file is a JSON blob that lists:

- the locations (IP/port) of the Prio servers and
- the name and type of each aggregate statistic the system is computing.

A number of [example config files are in the repository](https://github.com/henrycg/prio/tree/master/eval) but we will describe the relevant fields here.
What follows is an example configuration file:

    { 
      "maxPendingReqs": 64,
      "servers": [
        {"addrPub": "localhost:9000", "addrPriv": "localhost:9050"},
        {"addrPub": "localhost:9001", "addrPriv": "localhost:9051"},
        {"addrPub": "localhost:9002", "addrPriv": "localhost:9052"},
        {"addrPub": "localhost:9003", "addrPriv": "localhost:9053"},
        {"addrPub": "localhost:9004", "addrPriv": "localhost:9054"}
      ],
      "fields":[
        {"name": "hasUsedCamera", "type": "int", "intBits": 1}
        ,
        {"name": "hasUsedMicrophone", "type": "int", "intBits": 1}
        ,
        {"name": "cpuCores", "type": "int", "intBits": 4}
        ,
        {"name": "ramGB", "type": "int", "intBits": 6}
        ,
        {"name": "tabsOpenPerDay", "type": "int", "intBits": 12}
      ]
    }

To describe the meaning of each field 

- maxPendingReqs: determines how many client requests each server will 
  buffer. If you set this too large, the server will run out of memory.
  If you set this too small, the system will work but will be relatively slow.
  Set this to 64-128 and play with it to see what works.
- servers: a list of server IP/port combinations. Use "localhost" when
  using runservers to run a local deployment.

    - addrPub: the port on which _clients_ will connect to server
    - addrPriv: the port on which _servers_ will connect to the serer
- fields: a list of data fields. Each data field has:

    - name: a human-readable name
    - type: the data type. For now, the type "int" should be enough for
      most deployments. This privately sums up B-bit client-provided integers, 
      for your chosen value of B.
      The other data types are listed in [config/config.go](https://github.com/henrycg/prio/blob/master/config/config.go#L27) and you can see how to use them by looking at the [example config files](https://github.com/henrycg/prio/tree/master/eval). 
    - intBits: the largest integer value (in bits) 

If the config file above is saved as ~/myconfig.json, you could run

    $ cd $GOPATH/bin
    $ ./runservers -config ~/myconfig.json

to run a cluster of Prio servers using this configuration. To run
the corresponding client, run:

    $ ./tclient -config ~/myconfig.json

To start up the Prio servers manually (which is necessary if you want to run
them on different physical machines), you can run:

    $ ./tserver -config ~/myconfig.json -idx 3 -log /tmp/mylog.log

This starts up server 3 (indexed from 0) and configures
it to write logging information to /tmp/mylog.log.
The server will listen on the 3rd IP/port
combination listed in the configuation file.

