
import sys

def main():
    n = 0
    sizeTot = 0.0
    timeTot = 0.0
    for line in sys.stdin:
        if "Finished in " in line:
            n += 1
            timeTot += float(line.split()[4])
            print n
        if "Size" in line:
            sizeTot += float(line.split()[3])

    print "AvgSize: %lf" % (sizeTot/float(n))
    print "AvgTime: %lf" % (timeTot/float(n))

main()
