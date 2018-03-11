# google-home-notifier-go

Go version of [noelportugal/google-home-notifier](https://github.com/noelportugal/google-home-notifier)

## How to start

Run the following command.
You should set `GOOGLE_HOME_NOTIFIER_TOKEN`, if your port is opend to internet.

~~~
export GOOGLE_HOME_NOTIFIER_TOKEN=secrets
go build; ./google-home-notifier
~~~

## Test by curl

~~~
curl -XPOST "localhost:8080?text=てすと&lang=ja"&token=secrets"
~~~
