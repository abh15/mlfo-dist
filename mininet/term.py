#!/usr/bin/python3
import sys
from mininet.net import Containernet
from mininet.node import Controller, RemoteController, OVSController
from mininet.cli import CLI
from mininet.link import TCLink
from mininet.log import info, setLogLevel
from mininet.term import makeTerm
setLogLevel('info')

net = Containernet(controller=Controller)
info('*** Adding controller\n')
net.addController(name='c0',
                      controller=RemoteController,
                      ip='127.0.0.1',
                      protocol='tcp',
                      port=6633)
info('*** Adding docker containers\n')


numfognodes= int(sys.argv[1])
numedgeperfog = int(sys.argv[2])

#s0 is reserved for cloud connection
s0 = net.addSwitch("s0")
cloud0 = net.addDocker('cloud.0', ip='10.0.0.1', dimage="abh15/mlfo:latest",ports=[8000], port_bindings={8000:7000}, publish_all_ports=True)
cloud0.start()
net.addLink(cloud0, s0, cls=TCLink, delay='1ms', bw=0.5)


for i in range(1,numfognodes+1):
    fogsw = net.addSwitch("s"+ str(i))
    edgesw = net.addSwitch("s"+ str(i) + str(i))
    fognode = net.addDocker("fog."+str(i), ip="10.0."+str(i)+".1", dimage="abh15/mlfo:latest", cpu_period=50000, cpu_quota=25000)
    fognode.start()
    net.addLink(fognode, fogsw, cls=TCLink, delay='1ms',bw=0.5) #connect fog node to fog switch
    net.addLink(s0, fogsw, cls=TCLink, delay='1ms', bw=0.5) #connect fog switch to cloud switch
    net.addLink(fogsw, edgesw, cls=TCLink, delay='1ms', bw=0.5) #connect fog switch to edge switch
    for j in range(1, numedgeperfog+1):
        if i==1 and j==1:
            edgenode = net.addDocker("edge."+ str(i) + "." + str(j+1), ip="10.0."+ str(i) + "." + str(j+1), dimage="abh15/mlfo:latest", ports=[8000], port_bindings={8000:8000}, publish_all_ports=True)
        else:
            edgenode = net.addDocker("edge."+ str(i) + "." + str(j+1), ip="10.0."+ str(i) + "." + str(j+1), dimage="abh15/mlfo:latest")
        edgenode.start()
        net.addLink(edgenode, edgesw, cls=TCLink, delay='1ms', bw=0.5) #connect edge node to edge switch
    



info('*** Starting network\n')
net.start()
info('*** Running CLI\n')
CLI(net)
info('*** Stopping network')
net.stop()