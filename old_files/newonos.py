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


numfognodes= int(sys.argv[1])
numedgeperfog = int(sys.argv[2])

#s0 is reserved for cloud connection
s0 = net.addSwitch("s0",cls=OVSSwitch,protocols="OpenFlow13")
cloud0 = net.addDocker('cloud.0', ip='10.0.0.1', dimage="abh15/mlfo:latest",ports=[8000], port_bindings={8000:7000}, publish_all_ports=True)
cloud0.start()
net.addLink(cloud0, s0, cls=TCLink, delay='1ms', bw=10000)


for i in range(1,numfognodes+1):
    fogsw = net.addSwitch("s"+ str(i),cls=OVSSwitch,protocols="OpenFlow13")
    fognode = net.addDocker("fog."+str(i), ip="10.0."+str(i)+".1", dimage="abh15/mlfo:latest")#,cpu_period=100000, cpu_quota=99999)
    fognode.start()
    net.addLink(fognode, fogsw, cls=TCLink, delay='1ms',bw=10000) #connect fog node to fog switch
    net.addLink(s0, fogsw, cls=TCLink, delay='1ms', bw=10000) #connect fog switch to cloud switch
    for j in range(1, numedgeperfog+1):
        intentport = 8000+(i*10)+(j+1)
        edgenode = net.addDocker("edge."+ str(i) + "." + str(j+1), ip="10.0."+ str(i) + "." + str(j+1), dimage="abh15/mlfo:latest", ports=[8000], port_bindings={8000:intentport}, publish_all_ports=True)
        edgenode.start()
        #create 1 edgeswitch for every 10 edge nodes and connect to fogswitch
        if (j%10)==1:
            edgesw = net.addSwitch("s"+ str(j+10),cls=OVSSwitch,protocols="OpenFlow13")
            net.addLink(fogsw, edgesw, cls=TCLink, delay='1ms', bw=10000) #connect fog switch to edge switch
        net.addLink(edgenode, edgesw, cls=TCLink, delay='1ms', bw=10000) #connect edge node to edge switch

info('*** Starting network\n')
net.start()
info('*** Running CLI\n')
CLI(net)
info('*** Stopping network')
net.stop()