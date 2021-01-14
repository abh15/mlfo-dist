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
e3 = net.addDocker('edge.1.4', ip='10.0.1.4', dimage="abh15/mlfo:latest")
e4 = net.addDocker('edge.1.5', ip='10.0.1.5', dimage="abh15/mlfo:latest")
e5 = net.addDocker('edge.1.6', ip='10.0.1.6', dimage="abh15/mlfo:latest")

f2 = net.addDocker('fog.2', ip='10.0.2.1', dimage="abh15/mlfo:latest")
e6 = net.addDocker('edge.2.2', ip='10.0.2.2', dimage="abh15/mlfo:latest")
e7 = net.addDocker('edge.2.3', ip='10.0.2.3', dimage="abh15/mlfo:latest")
e8 = net.addDocker('edge.2.4', ip='10.0.2.4', dimage="abh15/mlfo:latest")
e9 = net.addDocker('edge.2.5', ip='10.0.2.5', dimage="abh15/mlfo:latest")
e10 = net.addDocker('edge.2.6', ip='10.0.2.6', dimage="abh15/mlfo:latest")

f3 = net.addDocker('fog.3', ip='10.0.3.1', dimage="abh15/mlfo:latest")
e11 = net.addDocker('edge.3.2', ip='10.0.3.2', dimage="abh15/mlfo:latest")
e12 = net.addDocker('edge.3.3', ip='10.0.3.3', dimage="abh15/mlfo:latest")
e13 = net.addDocker('edge.3.4', ip='10.0.3.4', dimage="abh15/mlfo:latest")
e14 = net.addDocker('edge.3.5', ip='10.0.3.5', dimage="abh15/mlfo:latest")
e15 = net.addDocker('edge.3.6', ip='10.0.3.6', dimage="abh15/mlfo:latest")

f4 = net.addDocker('fog.4', ip='10.0.4.1', dimage="abh15/mlfo:latest")
e16 = net.addDocker('edge.4.2', ip='10.0.4.2', dimage="abh15/mlfo:latest")
e17 = net.addDocker('edge.4.3', ip='10.0.4.3', dimage="abh15/mlfo:latest")
e18 = net.addDocker('edge.4.4', ip='10.0.4.4', dimage="abh15/mlfo:latest")
e19 = net.addDocker('edge.4.5', ip='10.0.4.5', dimage="abh15/mlfo:latest")
e20 = net.addDocker('edge.4.6', ip='10.0.4.6', dimage="abh15/mlfo:latest")

f5 = net.addDocker('fog.5', ip='10.0.5.1', dimage="abh15/mlfo:latest")
e21 = net.addDocker('edge.5.2', ip='10.0.5.2', dimage="abh15/mlfo:latest")
e22 = net.addDocker('edge.5.3', ip='10.0.5.3', dimage="abh15/mlfo:latest")
e23 = net.addDocker('edge.5.4', ip='10.0.5.4', dimage="abh15/mlfo:latest")
e24 = net.addDocker('edge.5.5', ip='10.0.5.5', dimage="abh15/mlfo:latest")
e25 = net.addDocker('edge.5.6', ip='10.0.5.6', dimage="abh15/mlfo:latest")

cloud0 = net.addDocker('cloud.0', ip='10.0.0.1', dimage="abh15/mlfo:latest")

info('*** Adding switches\n')
s1 = net.addSwitch('s1')
s2 = net.addSwitch('s2')
s3 = net.addSwitch('s3')
s4 = net.addSwitch('s4')
s5 = net.addSwitch('s5')
s6 = net.addSwitch('s6')

info('*** Creating links\n')
net.addLink(cloud0, s1, delay='1ms')

net.addLink(f1, s2, delay='2ms')
net.addLink(e1, s2, delay='2ms')
net.addLink(e2, s2, delay='2ms')
net.addLink(e3, s2, delay='2ms')
net.addLink(e4, s2, delay='2ms')
net.addLink(e5, s2, delay='2ms')

net.addLink(f2, s3, delay='2ms')
net.addLink(e6, s3, delay='2ms')
net.addLink(e7, s3, delay='2ms')
net.addLink(e8, s3, delay='2ms')
net.addLink(e9, s3, delay='2ms')
net.addLink(e10, s3, delay='2ms')

net.addLink(f3, s4, delay='2ms')
net.addLink(e11, s4, delay='2ms')
net.addLink(e12, s4, delay='2ms')
net.addLink(e13, s4, delay='2ms')
net.addLink(e14, s4, delay='2ms')
net.addLink(e15, s4, delay='2ms')

net.addLink(f4, s5, delay='2ms')
net.addLink(e16, s5, delay='2ms')
net.addLink(e17, s5, delay='2ms')
net.addLink(e18, s5, delay='2ms')
net.addLink(e19, s5, delay='2ms')
net.addLink(e20, s5, delay='2ms')


net.addLink(f5, s6, delay='2ms')
net.addLink(e21, s6, delay='2ms')
net.addLink(e22, s6, delay='2ms')
net.addLink(e23, s6, delay='2ms')
net.addLink(e24, s6, delay='2ms')
net.addLink(e25, s6, delay='2ms')



net.addLink(s1, s2, cls=TCLink, delay='40ms')
net.addLink(s1, s3, cls=TCLink, delay='40ms')
net.addLink(s1, s4, cls=TCLink, delay='40ms')
net.addLink(s1, s5, cls=TCLink, delay='40ms')
net.addLink(s1, s6, cls=TCLink, delay='40ms')

info('*** Starting network\n')
net.start()
info('*** Running CLI\n')
CLI(net)
info('*** Stopping network')
net.stop()
