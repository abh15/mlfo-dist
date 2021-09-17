#!/usr/bin/python3
#Script to simulate LEO satellite topo for a industrial edge

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

numrobots = int(sys.argv[1])

fserver1 = net.addDocker('fed.1', ip='10.0.0.101', dimage="abh15/flwr:latest") 
fserver1.start()
fserver2 = net.addDocker('fed.2', ip='10.0.0.102', dimage="abh15/flwr:latest") 
fserver2.start()
cloud0 = net.addDocker('cloud.0', ip='10.0.0.1', dimage="abh15/mlfo:latest",ports=[8000], port_bindings={8000:8999}, publish_all_ports=True) 
cloud0.start()

aggsw = net.addSwitch("aggsw0",cls=OVSSwitch,protocols="OpenFlow13")
LEOsw = net.addSwitch("satsw0",cls=OVSSwitch,protocols="OpenFlow13")

net.addLink(cloud0, aggsw)  
net.addLink(fserver1, aggsw)
net.addLink(fserver2, aggsw)
net.addLink(LEOsw, aggsw, cls=TCLink, delay="12ms", bw=100)  

#===========================================================================

intentport = 8000+1
edgesw = net.addSwitch("swEdge0",cls=OVSSwitch,protocols="OpenFlow13")
mlfonode = net.addDocker("mo.1", ip="10.0.1.1", dimage="abh15/mlfo:latest", ports=[8000], port_bindings={8000:intentport}, publish_all_ports=True) 
mlfonode.start()
net.addLink(mlfonode, edgesw)
for j in range(1, numrobots+1):
    fclient = net.addDocker("fc."+ str(j+10), ip="10.0.1." + str(j+10), dimage="abh15/flwr:latest") 
    fclient.start()
    net.addLink(fclient, edgesw) 
net.addLink(edgesw, LEOsw, cls=TCLink, loss=3, delay="12ms", bw=100)
 

info('*** Starting network\n')
net.start()
info('*** Running CLI\n')
CLI(net)
info('*** Stopping network')
net.stop()