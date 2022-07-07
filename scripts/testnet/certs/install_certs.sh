
sudo apk add certbot
sudo apk add python3 python3-dev py3-pip

sudo pip install google-cloud-dns
sudo pip install certbot-dns-google

sudo certbot certonly --dns-google --dns-google-credentials terraform.json \
   -d *.testnet.stridenet.co \
   -d *.poolparty.stridenet.co \
   -d *.internal.stridenet.co \
   -d *.aidan.stridenet.co \
   -d *.sam.stridenet.co \
   -d *.riley.stridenet.co \
   -d *.vishal.stridenet.co
 