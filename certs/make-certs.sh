if [[ ! -f root-ca-key.pem ]]; then
  echo "gen root"
  openssl genrsa -out root-ca-key.pem 2048
  openssl req -new -x509 -sha256 -key root-ca-key.pem -subj "/C=UK/L=London/O=foo/CN=haproy" -out root-ca.pem -days 730
fi
echo "gen haproy"
openssl genrsa -out haproxy-key-temp.pem 2048
openssl pkcs8 -inform PEM -outform PEM -in haproxy-key-temp.pem -topk8 -nocrypt -v1 PBE-SHA1-3DES -out haproxy-key.pem
openssl req -new -key haproxy-key.pem -subj "/C=UK/L=London/O=foo/CN=haproy" -out haproxy.csr
openssl x509 -req -in haproxy.csr -CA root-ca.pem -CAkey root-ca-key.pem -CAcreateserial -sha256 -out haproxy.pem -days 730


echo """
# Specify PEM in haproxy config
sudo vim /etc/haproxy/haproxy.cfg
listen haproxy
  bind 0.0.0.0:443 ssl crt /etc/ssl/private/mydomain.pem
"""
