import urllib2
from v2 import *
import json
import sys
import requests

C_APPNAME = "client.py"
C_CREDAPIURL = "http://127.0.0.1:8889/"

# 1. get key
if(len(sys.argv)<1):
	sys.exit(0)

credid = sys.argv[1]
keyid = ""

# 1. get creds and associated key id
#r = requests.post(C_CREDAPIURL + "cred", data = {"Credid":credid, "Appname":C_APPNAME})

payload = {'Credid':credid,'Appname':C_APPNAME}
response = requests.post(C_CREDAPIURL + "cred", data=json.dumps(payload))
data = response.json()

try:
	secretinfo = data['secretinfo']
	keyid = data['keyid']

except KeyError:
	    print("ERROR: Could not retrieve key id from cred request")

print(" >> SECRETINFO: " + secretinfo)
print(" >> KEYID: " + keyid)

# 2. get key
payload = {'Keyid':keyid,'Appname':C_APPNAME}

response = requests.post(C_CREDAPIURL + "key", data=json.dumps(payload))
data = response.json()

key = data['key']
print(" >> KEY: " + key)

#jsonresponse = response.read()
#print jsonresponse
#data = json.loads(jsonresponse)
#key = data['key']
#print "Key: " + key

# 2. get credentials from database
#response = urllib2.urlopen(C_CREDAPIURL + "cred?credid=" + credid + "&appname=client.py")
#response = requests.post(C_CREDAPIURL + "key", data = {"credid":credid, "appname":C_APPNAME})

#jsonresponse = response.read()
#data = json.loads(jsonresponse)
#ciphertext = data['secretinfo']
#print "Ciphertext: " + ciphertext

#secretinfo = "somepassword"
#key = "USWASEUTNEBS"
decr = decrypt(secretinfo, key)
print "Plaintext: " + decr

# ========================
enc = encrypt("REVEALED","NESWABKBQMGOWDYIOKRE")
print enc