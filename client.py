import urllib2
from v2 import *
import json
import sys

scriptName = "client.py"
resourceName = "vertica1"
C_CREDAPIURL = "http://127.0.0.1:8888/"

# 1. get key
if(len(sys.argv)<2):
	sys.exit(0)

keyid = sys.argv[1]
credid = sys.argv[2]
print "Key ID: " + keyid
print "Cred ID: " + credid

response = urllib2.urlopen(C_CREDAPIURL + "key?keyid=" + keyid + "&appname=client.py")
jsonresponse = response.read()
data = json.loads(jsonresponse)
#print data['key']
key = data['key']
print "Key: " + key

# 2. get credentials from database
response = urllib2.urlopen(C_CREDAPIURL + "cred?credid=" + credid + "&appname=client.py")
jsonresponse = response.read()
data = json.loads(jsonresponse)
ciphertext = data['secretinfo']
print "Ciphertext: " + ciphertext

decr = decrypt(ciphertext, key)
print "Plaintext: " + decr

