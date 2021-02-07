#!/usr/bin/python3
import sys
from mininet.net import Containernet
from mininet.node import Controller
from mininet.cli import CLI
from mininet.link import TCLink
from mininet.log import info, setLogLevel
setLogLevel('info')

net = Containernet(controller=Controller)
info('*** Adding controller\n')
net.addController('c0')
info('*** Adding docker containers\n')


numfognodes= int(sys.argv[1])
numedgeperfog = int(sys.argv[2])

#s0 is reserved for cloud connection
s0 = net.addSwitch("s0")
cloud0 = net.addDocker('cloud.0', ip='10.0.0.1', dimage="abh15/mlfo:latest")
cloud0.start()
net.addLink(cloud0, s0, delay='1ms')


for i in range(1,numfognodes+1):
    sw = net.addSwitch("s"+ str(i))
    fognode = net.addDocker("fog."+str(i), ip="10.0."+str(i)+".1", dimage="abh15/mlfo:latest", cpu_period=50000, cpu_quota=10000)
    fognode.start()
    net.addLink(fognode, sw, delay='1ms')
    for j in range(1, numedgeperfog+1):
        if i==1 and j==1:
            edgenode = net.addDocker("edge."+ str(i) + "." + str(j+1), ip="10.0."+ str(i) + "." + str(j+1), dimage="abh15/mlfo:latest", ports=[8000], port_bindings={8000:8000}, publish_all_ports=True)
        else:
            edgenode = net.addDocker("edge."+ str(i) + "." + str(j+1), ip="10.0."+ str(i) + "." + str(j+1), dimage="abh15/mlfo:latest")
        edgenode.start()
        net.addLink(edgenode, sw, delay='1ms')
    net.addLink(s0, sw, cls=TCLink, delay='1ms') #sw-sw links to cloud switch



info('*** Starting network\n')
net.start()
info('*** Running CLI\n')
CLI(net)
info('*** Stopping network')
net.stop()
