#!/usr/bin/python3
import asyncio
import aiohttp
import sys
from aiohttp import FormData

async def get(url, session):
    try:
        payload = FormData()
        payload.add_field('file',  open('intent.yaml', 'rb'))
        payload.add_field('nodesperedge', sys.argv[2])
        payload.add_field('clipernode', sys.argv[3])
        payload.add_field('mlfostatus', sys.argv[4])
        payload.add_field('flstatus', sys.argv[5])
        payload.add_field('hierflstatus', sys.argv[6])
        async with session.post(url, data=payload) as resp:
            print(resp.status)
            print(await resp.text())
    except Exception as e:
        print("Unable to get url {} due to {}.".format(url, e.__class__))


async def main(urls):
    async with aiohttp.ClientSession() as session:
        ret = await asyncio.gather(*[get(url, session) for url in urls])
    print("Finalized all. Return is a list of len {} outputs.".format(len(ret)))

urls= []

for i in range(1,(int(sys.argv[1])+1)):
    port = 8000+i
    urls.append("http://10.66.2.142:"+ str(port) + "/receive")

# print (urls)
asyncio.run(main(urls))
