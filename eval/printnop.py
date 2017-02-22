import sys
S = 5#int(sys.argv[1])
np = int(sys.argv[1])
N = np * (np+1)/2

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
""" % ( 262144/N, ",\n".join(servers[0:S]))


for i in range(N):
    print '{"name": "cell%d", "type": "intUnsafe", "intBits": 1}' % i
    if (i+1) != N: 
        print ","
#print '{"name": "batch", "type": "intBatch", "intBatchLen": %d}' % N
print "]}"
