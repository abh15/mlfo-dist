#!/usr/bin/python3
import requests
import sys
import time

"""" 
IP range 10.0.1.10 to 10.0.1.59
Run python3 sendintentpy. 
Pls check length of all arrays.
If sameserverdist is ues then that cohort shared fedavg server with the cohort before that one
"""

#timegap in seconds is period in between two intents 
timegap=1


# cohortdistr = [6,6,6,6,1,1,1,1,1,1]
# intentdist=['intent.yaml','2intent.yaml','intent.yaml','2intent.yaml','intent.yaml', '2intent.yaml', 'intent.yaml', '2intent.yaml', 'intent.yaml', '2intent.yaml','intent.yaml', '2intent.yaml']
# # intentdist=['3intent.yaml','3intent.yaml','3intent.yaml','3intent.yaml','3intent.yaml','3intent.yaml','3intent.yaml','3intent.yaml']
# sameserverdist= ['nor','nor','nor','nor','nor','nor','nor','nor','nor','nor']
# avgalgodist=['FedAvg','FedAvg','FedAvg','FedAvg','FedAvg','FedAvg','FedAvg','FedAvg','FedAvg','FedAvg']
# fracfitdist= ['0.5','0.5','0.5','0.5','0.5','0.5','0.5','0.5','0.5','0.5']
# minfitdist=['1','1','1','1','1','1','1','1','1','1']
# minavdist= ['1','1','1','1','1','1','1','1','1','1']
# numrounddist =['20','20','20','20','20','20','20','20','20','20']




def send(intentfile, ipstart, cohortsize, sameserver, avgalgo, fracfit, minfit, minav, numround):
    files = {'file': open(intentfile, 'rb')}
    payload = {'ipstart': ipstart, 'cohortsize' : cohortsize, 'sameserver': sameserver, 'avgalgo': avgalgo,
                'fracfit':fracfit, 'minfit':minfit, 'minav':minav, 'numround':numround}
    r = requests.post('http://10.66.2.142:8001/receive', files=files, data=payload)
       
def main():
    octtraker = 10
    for i in range(30):
        cohort= 1
        intent= 'intent.yaml'
        sameserver= 'nor'
        avgalgo= 'FedAvg'
        fracfit= '0.5'
        minfit= '1'
        minav= '1'
        numround= '20'
        send(intent, octtraker, cohort, sameserver, avgalgo, fracfit, minfit, minav, numround)
        octtraker=octtraker+cohort
        time.sleep(timegap)


main()
