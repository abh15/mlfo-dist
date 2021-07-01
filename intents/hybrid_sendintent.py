#!/usr/bin/python3
# only one input arg- total number of edges(must be even)
import asyncio
import aiohttp
import sys
from aiohttp import FormData

async def getwithmlfo(url, session):
    try:
        payload = FormData()
        payload.add_field('file',  open('intent.yaml', 'rb'))
        payload.add_field('nodesperedge', "3")
        payload.add_field('clipernode', "3")
        payload.add_field('mlfostatus', "enabled")
        payload.add_field('flstatus', "disabled")
        payload.add_field('hierflstatus', "disabled")
        async with session.post(url, data=payload) as resp:
            print(resp.status)
            print(await resp.text())
    except Exception as e:
        print("Unable to get url {} due to {}.".format(url, e.__class__))



async def getwomlfo(url, session):
    try:
        payload = FormData()
        payload.add_field('file',  open('intent.yaml', 'rb'))
        payload.add_field('nodesperedge', "3")
        payload.add_field('clipernode', "3")
        payload.add_field('mlfostatus', "disabled")
        payload.add_field('flstatus', "disabled")
        payload.add_field('hierflstatus', "disabled")
        async with session.post(url, data=payload) as resp:
            print(resp.status)
            print(await resp.text())
    except Exception as e:
        print("Unable to get url {} due to {}.".format(url, e.__class__))



async def main(urlswithmlfo, urlswomlfo):
    async with aiohttp.ClientSession() as session:
        ret = await asyncio.gather(*[getwithmlfo(url, session) for url in urlswithmlfo])
    print("Finalized all. Return is a list of len {} outputs.".format(len(ret)))
    async with aiohttp.ClientSession() as session:
        ret = await asyncio.gather(*[getwomlfo(url, session) for url in urlswomlfo])
    print("Finalized all. Return is a list of len {} outputs.".format(len(ret)))



urlswithmlfo= []
urlswomlfo= []
num = int(sys.argv[1])
flag = True
for i in range(1,num+1):
    if flag==True:
        port = 8000+i
        urlswithmlfo.append("http://10.66.2.142:"+ str(port) + "/receive")
    else:
        port = 8000+i
        urlswomlfo.append("http://10.66.2.142:"+ str(port) + "/receive")

    if i%(num/6)==0:
        flag= not flag
    
print(urlswithmlfo)
print(urlswomlfo)



   










# print (urls)
asyncio.run(main(urlswithmlfo, urlswomlfo))
