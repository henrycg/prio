import sys
S = 5#int(sys.argv[1])
N = int(sys.argv[1])

ips = [
'54.226.40.248',
'35.167.90.45',
'54.153.87.79',
'35.157.55.160',
'34.249.119.221',
]
servers = [ '{"addrPub": "%s:%d", "addrPriv": "%s:%d"}' % (s, 9000+i, s, 9050+i) for (i,s) in enumerate(ips) ] 
print """
{ "maxPendingReqs": %d,
"clients": [ "54.196.12.191", "184.72.117.74", "184.72.120.112"],
"servers": [
%s
],
"fields":[
      {"name": "linReg0", "type": "linReg",
          "linRegBits": [%s]}
]
}
""" % ( 1024*8/N, ",\n".join(servers[0:S]), ",".join(["14"]*N))
