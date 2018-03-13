# google-home-notifier-go

Go version of [noelportugal/google-home-notifier](https://github.com/noelportugal/google-home-notifier)

## How to start

Run the following command.
You should set `GHN_TOKEN`, if your port is opend to internet.

~~~
export GHN_TOKEN=secrets
export GHN_PORT=80
go build; ./google-home-notifier
~~~

## Start with Docker

~~~
docker run -d -e GHN_TOKEN=secrets -e GHN_PORT=80 --net=host juntaki/google-home-notifier-go
~~~

## Test by curl

~~~
curl -XPOST "localhost:8080?text=てすと&lang=ja"&token=secrets"
~~~
