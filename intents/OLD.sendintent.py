#!/usr/bin/python3
import asyncio
import aiohttp
import sys
from aiohttp import FormData

async def get(url, session):
    try:
        payload = FormData()
        payload.add_field('file',  open('intent.yaml', 'rb'))
        payload.add_field('ipstart', "10")
        payload.add_field('cohortsize', "5")
        payload.add_field('sameserver', "no")
        payload.add_field('avgalgo', "FedAvg")
        payload.add_field('fracfit', "0.5")
        payload.add_field('minfit', "1")
        payload.add_field('minav', "1")
        payload.add_field('numround', "5")
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
