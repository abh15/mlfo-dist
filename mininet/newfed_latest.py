#!/usr/bin/python3
import sys
from mininet.net import Containernet
from mininet.node import Controller, RemoteController, OVSSwitch
from mininet.cli import CLI
from mininet.link import TCLink
from mininet.log import info, setLogLevel

setLogLevel('info')

net = Containernet(controller=Controller, switch=OVSSwitch)
info('*** Adding controller\n')

net.addController(name='c0', controller=RemoteController, ip='127.0.0.1', port=6653, protocols="OpenFlow13")
info('*** Adding docker containers\n')

numsat = int(sys.argv[1])
numfwa = int(sys.argv[2])
nummet = int(sys.argv[3])
numrobots = int(sys.argv[4])

fserver1 = net.addDocker('fed.1', ip='10.0.0.101', dimage="abh15/flwr:latest") 
fserver1.start()
fserver2 = net.addDocker('fed.2', ip='10.0.0.102', dimage="abh15/flwr:latest") 
fserver2.start()
cloud0 = net.addDocker('cloud.0', ip='10.0.0.1', dimage="abh15/mlfo:latest",ports=[8000], port_bindings={8000:8999}, publish_all_ports=True) 
cloud0.start()

aggsw = net.addSwitch("aggs0",cls=OVSSwitch,protocols="OpenFlow13")
GSTAsw = net.addSwitch("gstas0",cls=OVSSwitch,protocols="OpenFlow13")
FWAsw = net.addSwitch("fwas0",cls=OVSSwitch,protocols="OpenFlow13")
METsw = net.addSwitch("mets0",cls=OVSSwitch,protocols="OpenFlow13")


net.addLink(cloud0, aggsw, cls=TCLink, delay='2.5ms',bw=10000)  
net.addLink(fserver1, aggsw, cls=TCLink, delay='2.5ms',bw=10000)
net.addLink(fserver2, aggsw, cls=TCLink, delay='2.5ms',bw=10000)

net.addLink(GSTAsw, aggsw, cls=TCLink, delay='2.5ms',bw=10000)  
net.addLink(FWAsw, aggsw, cls=TCLink, delay='2.5ms',bw=10000)
net.addLink(METsw, aggsw, cls=TCLink, delay='2.5ms',bw=10000)

def subtopo (start_num, end_num, delay, bw, pop_sw, prefix):
    for i in range(start_num,end_num+1):
        intentport = 8000+(i)
        campussw = net.addSwitch("s"+ str(i),cls=OVSSwitch,protocols="OpenFlow13")
        mlfonode = net.addDocker(prefix+"mo."+str(i), ip="10.0."+str(i)+".1", dimage="abh15/mlfo:latest", ports=[8000], port_bindings={8000:intentport}, publish_all_ports=True) 
        mlfonode.start()
        net.addLink(mlfonode, campussw, cls=TCLink, delay=delay,bw=bw)
        net.addLink(campussw, pop_sw, cls=TCLink, delay=delay, bw=bw) 
        flocalagg = net.addDocker("fla."+ str(i), ip="10.0."+ str(i) + ".100", dimage="abh15/flwr:latest")
        flocalagg.start()
        net.addLink(flocalagg, campussw, cls=TCLink, delay=delay, bw=bw)
        for j in range(1, numrobots+1):
            fclient = net.addDocker("fc."+ str(i) + str(j+10), ip="10.0."+ str(i) + "." + str(j+10), dimage="abh15/flwr:latest") 
            fclient.start()
            net.addLink(fclient, campussw, cls=TCLink, delay=delay, bw=bw) 

 

subtopo(1, numsat, "15ms", 600, GSTAsw, "s")

subtopo(numsat+1, (numsat+numfwa), "10ms", 2000, FWAsw, "f")

subtopo((numsat+numfwa+1), (numsat+numfwa+nummet), "5ms", 10000, METsw, "m")



info('*** Starting network\n')
net.start()
info('*** Running CLI\n')
CLI(net)
info('*** Stopping network')
net.stop()