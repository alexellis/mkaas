FROM alpine:latest

RUN apk --update add squid

# This step enables access from private IP address ranges.
RUN sed -ie s/"http_access deny all"/"http_access allow all"/g /etc/squid/squid.conf
RUN sed -ie s/"http_access deny !Safe_ports"/"http_access allow !Safe_ports"/g /etc/squid/squid.conf
RUN sed -ie s/"http_access deny CONNECT !SSL_ports"/"http_access allow CONNECT !SSL_ports"/g /etc/squid/squid.conf

CMD ["squid", "-NYCd", "1"]
