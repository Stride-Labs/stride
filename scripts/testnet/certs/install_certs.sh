
sudo apk add certbot
sudo apk add python3 python3-dev py3-pip

sudo pip install google-cloud-dns
sudo pip install certbot-dns-google

sudo certbot certonly --dns-google --dns-google-credentials terraform.json \
   -d *.poolparty.stridenet.co \
   -d *.frontend.stridenet.co \
   -d *.staging.stridenet.co \
   -d *.testnet-1.stridenet.co \
   -d *.testnet-2.stridenet.co \
   -d *.testnet-3.stridenet.co \
   -d *.testnet-4.stridenet.co
 
 sudo certbot certonly --dns-google --dns-google-credentials terraform.json \
   -d *.poolparty.stridenet.co 
