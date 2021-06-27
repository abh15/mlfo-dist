# #!/usr/bin/python3
# import sys
# from mininet.net import Containernet
# from mininet.node import Controller, RemoteController, OVSSwitch
# from mininet.cli import CLI
# from mininet.link import TCLink
# from mininet.log import info, setLogLevel

# setLogLevel('info')

# net = Containernet(controller=Controller, switch=OVSSwitch)
# info('*** Adding controller\n')

# net.addController(name='c0', controller=RemoteController, ip='127.0.0.1', port=6653, protocols="OpenFlow13")
# info('*** Adding docker containers\n')

# numsat = int(sys.argv[1])
# numfwa = int(sys.argv[2])
# nummet = int(sys.argv[3])
# numrobots = int(sys.argv[4])

# fserver1 = net.addDocker('fed.1', ip='10.0.0.101', dimage="abh15/flwr:latest") 
# fserver1.start()
# fserver2 = net.addDocker('fed.2', ip='10.0.0.102', dimage="abh15/flwr:latest") 
# fserver2.start()
# cloud0 = net.addDocker('cloud.0', ip='10.0.0.1', dimage="abh15/mlfo:latest",ports=[8000], port_bindings={8000:8999}, publish_all_ports=True) 
# cloud0.start()

# aggsw = net.addSwitch("aggs0",cls=OVSSwitch,protocols="OpenFlow13")
# GSTAsw = net.addSwitch("gstas0",cls=OVSSwitch,protocols="OpenFlow13")
# FWAsw = net.addSwitch("fwas0",cls=OVSSwitch,protocols="OpenFlow13")
# METsw = net.addSwitch("mets0",cls=OVSSwitch,protocols="OpenFlow13")


# net.addLink(cloud0, aggsw, cls=TCLink, delay='1ms', bw=10000)  
# net.addLink(fserver1, aggsw, cls=TCLink, delay='1ms', bw=10000)
# net.addLink(fserver2, aggsw, cls=TCLink, delay='1ms', bw=10000)

# net.addLink(GSTAsw, aggsw, cls=TCLink, delay='1ms', bw=10000)  
# net.addLink(FWAsw, aggsw, cls=TCLink, delay='1ms', bw=10000)
# net.addLink(METsw, aggsw, cls=TCLink, delay='1ms', bw=10000)


# for i in range(1,(numsat+1)):
#     intentport = 8000+(i)
#     edgesw = net.addSwitch("s"+ str(i),cls=OVSSwitch,protocols="OpenFlow13")
#     mlfonode = net.addDocker("smo."+str(i), ip="10.0."+str(i)+".1", dimage="abh15/mlfo:latest", ports=[8000], port_bindings={8000:intentport}, publish_all_ports=True) 
#     mlfonode.start()
#     net.addLink(mlfonode, edgesw, cls=TCLink, delay='1ms', bw=10000)
#     flocalagg = net.addDocker("fla."+ str(i), ip="10.0."+ str(i) + ".100", dimage="abh15/flwr:latest")
#     flocalagg.start()
#     net.addLink(flocalagg, edgesw, cls=TCLink, delay='1ms', bw=10000)
#     for j in range(1, numrobots+1):
#         fclient = net.addDocker("fc."+ str(i) + str(j+10), ip="10.0."+ str(i) + "." + str(j+10), dimage="abh15/flwr:latest") 
#         fclient.start()
#         net.addLink(fclient, edgesw, cls=TCLink, delay='1ms', bw=10000) 
#     net.addLink(edgesw, GSTAsw, cls=TCLink, delay='1ms', bw=10000) 

# for i in range((numsat+1),(numsat+numfwa+1)):
#     intentport = 8000+(i)
#     edgesw = net.addSwitch("s"+ str(i),cls=OVSSwitch,protocols="OpenFlow13")
#     mlfonode = net.addDocker("fmo."+str(i), ip="10.0."+str(i)+".1", dimage="abh15/mlfo:latest", ports=[8000], port_bindings={8000:intentport}, publish_all_ports=True) 
#     mlfonode.start()
#     net.addLink(mlfonode, edgesw, cls=TCLink, delay='1ms', bw=10000)
#     flocalagg = net.addDocker("fla."+ str(i), ip="10.0."+ str(i) + ".100", dimage="abh15/flwr:latest")
#     flocalagg.start()
#     net.addLink(flocalagg, edgesw, cls=TCLink, delay='1ms', bw=10000)
#     for j in range(1, numrobots+1):
#         fclient = net.addDocker("fc."+ str(i) + str(j+10), ip="10.0."+ str(i) + "." + str(j+10), dimage="abh15/flwr:latest") 
#         fclient.start()
#         net.addLink(fclient, edgesw, cls=TCLink, delay='1ms', bw=10000) 
#     net.addLink(edgesw, FWAsw, cls=TCLink, delay='1ms', bw=10000) 


# for i in range((numsat+numfwa+1),(numsat+numfwa+nummet+1)):
#     intentport = 8000+(i)
#     edgesw = net.addSwitch("s"+ str(i),cls=OVSSwitch,protocols="OpenFlow13")
#     mlfonode = net.addDocker("mmo."+str(i), ip="10.0."+str(i)+".1", dimage="abh15/mlfo:latest", ports=[8000], port_bindings={8000:intentport}, publish_all_ports=True) 
#     mlfonode.start()
#     net.addLink(mlfonode, edgesw, cls=TCLink, delay='1ms', bw=10000)
#     flocalagg = net.addDocker("fla."+ str(i), ip="10.0."+ str(i) + ".100", dimage="abh15/flwr:latest")
#     flocalagg.start()
#     net.addLink(flocalagg, edgesw, cls=TCLink, delay='1ms', bw=10000)
#     for j in range(1, numrobots+1):
#         fclient = net.addDocker("fc."+ str(i) + str(j+10), ip="10.0."+ str(i) + "." + str(j+10), dimage="abh15/flwr:latest") 
#         fclient.start()
#         net.addLink(fclient, edgesw, cls=TCLink, delay='1ms', bw=10000) 
#     net.addLink(edgesw, METsw, cls=TCLink, delay='1ms', bw=10000)     
 

# info('*** Starting network\n')
# net.start()
# info('*** Running CLI\n')
# CLI(net)
# info('*** Stopping network')
# net.stop()


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

cloud0 = net.addDocker('cloud.0', ip='10.0.0.1', dimage="abh15/mlfo:latest",ports=[8000], port_bindings={8000:8999}, publish_all_ports=True)
cloud0.start()

aggsw = net.addSwitch("aggs0",cls=OVSSwitch,protocols="OpenFlow13")
FWAsw = net.addSwitch("fwas0",cls=OVSSwitch,protocols="OpenFlow13")

net.addLink(cloud0, aggsw, cls=TCLink, delay='1ms', bw=10000)
net.addLink(FWAsw, aggsw, cls=TCLink, delay='1ms', bw=10000)

i=1


intentport = 8000+(i)
edgesw = net.addSwitch("s"+ str(i),cls=OVSSwitch,protocols="OpenFlow13")
mlfonode = net.addDocker("fmo."+str(i), ip="10.0."+str(i)+".1", dimage="abh15/mlfo:latest", ports=[8000], port_bindings={8000:intentport}, publish_all_ports=True)
mlfonode.start()
net.addLink(mlfonode, edgesw, cls=TCLink, delay='1ms', bw=10000)
net.addLink(edgesw, FWAsw, cls=TCLink, delay='1ms', bw=10000)

info('*** Starting network\n')
net.start()
info('*** Running CLI\n')
CLI(net)
info('*** Stopping network')
net.stop()
