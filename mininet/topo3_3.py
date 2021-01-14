#!/usr/bin/python
"""
This is the most simple example to showcase Containernet.
"""
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

f1 = net.addDocker('fog.1', ip='10.0.1.1', dimage="abh15/mlfo:latest")
e1 = net.addDocker('edge.1.2', ip='10.0.1.2', dimage="abh15/mlfo:latest", ports=[8000], port_bindings={8000:8000}, publish_all_ports=True)
e2 = net.addDocker('edge.1.3', ip='10.0.1.3', dimage="abh15/mlfo:latest")

f2 = net.addDocker('fog.2', ip='10.0.2.1', dimage="abh15/mlfo:latest")
e3 = net.addDocker('edge.2.2', ip='10.0.2.2', dimage="abh15/mlfo:latest")
e4 = net.addDocker('edge.2.3', ip='10.0.2.3', dimage="abh15/mlfo:latest")

cloud0 = net.addDocker('cloud.0', ip='10.0.0.1', dimage="abh15/mlfo:latest")

info('*** Adding switches\n')
s1 = net.addSwitch('s1')
s2 = net.addSwitch('s2')
s3 = net.addSwitch('s3')

info('*** Creating links\n')
net.addLink(cloud0, s1)

net.addLink(f1, s2)
net.addLink(e1, s2,)
net.addLink(e2, s2)

net.addLink(f2, s3)
net.addLink(e3, s3)
net.addLink(e4, s3)

net.addLink(s1, s2)
net.addLink(s1, s3)

info('*** Starting network\n')
net.start()
info('*** Running CLI\n')
CLI(net)
info('*** Stopping network')
net.stop()
