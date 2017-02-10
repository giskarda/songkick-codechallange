# songkick-codechallange


##Requirements:

``go get github.com/patrickmn/go-cache``

## Build

``go build songkick.go``

##Usage:

### Server:
Make sure to run:
``firewall.sh``

And after that simply run:

``./songkick``

###Client:

curl -H 'Host: songkick-api-proxy' "localhost:8080/api/3.0/search/artists.json?query=muse&apikey=TkHqXOx7ZOhtT69x"

Host: is mandatory and it has to be that period.
