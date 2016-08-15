import urllib2
response = urllib2.urlopen('http://127.0.0.1:8888/find?id=2&req=VIO28FYT5S9Y6LP3G70PPC6XQV0W5R1GBC7VNZ3KY7Z5XPN94R9FBB1VSS9ZXBG6')
html = response.read()
print html