# Go Rest Client

gorestclient provides the functionality to perform GET, POST, PUT, DELETE and HEAD in a very simple way.
 
## Simple Usage
The most basic usage of the gorestclient is the following

    //Do the GET call to ML API
	response, err := Get("https://api.mercadolibre.com/sites/MLA")

	//Do the POST call
	response, err := Post("https://api.mercadolibre.com/sites", "{\"id\":\"MLA\"}")

	//Do the PUT call
	response, err := Put("https://api.mercadolibre.com/sites/MLA", "{\"id\":\"MLA\"}")

	//Do the DELETE call
	response, err := Delete("https://api.mercadolibre.com/sites/MLA")

	//Do the HEAD call
	response, err := Head("https://api.mercadolibre.com/sites/MLA")

## Add Headers
If you want to add some headers to your calls, you can do:

    //Do the GET call to ML API
	response, err := Get("https://api.mercadolibre.com/sites/MLA", Header{Key: "Test", Value: "Test Header"}, Header{Key: "Other", Value: "Other Test Header"})

	//Do the POST call
	response, err := Post("https://api.mercadolibre.com/sites", "{\"id\":\"MLA\"}", Header{Key: "Test", Value: "Test Header"}, Header{Key: "Other", Value: "Other Test Header"})

	//Do the PUT call
	response, err := Put("https://api.mercadolibre.com/sites/MLA", "{\"id\":\"MLA\"}", Header{Key: "Test", Value: "Test Header"}, Header{Key: "Other", Value: "Other Test Header"})

	//Do the DELETE call
	response, err := Delete("https://api.mercadolibre.com/sites/MLA", Header{Key: "Test", Value: "Test Header"}, Header{Key: "Other", Value: "Other Test Header"})

	//Do the HEAD call
	response, err := Head("https://api.mercadolibre.com/sites/MLA", Header{Key: "Test", Value: "Test Header"}, Header{Key: "Other", Value: "Other Test Header"})

## Configuration
If you not configure any connection pool, a default conection pool will be used, but in case that you want
to customize the connection pool you can do it in a very easy way:

	//Create a custom config
	config := new(PoolConfig)
	config.BaseURL = "http://internal.mercadolibre.com"
	config.MaxIdleConnsPerHost = 20
	config.Timeout = 100
	config.CacheElements = 100
    config.CacheStale = true
    config.Proxy = "http://183.123.334.222:8080"
    
    //Add the customization to the connection pool
	AddCustomPool("/sites/.*", config)

    //Do the GET call to ML API
	response, err := Get("/sites/MLA")

BaseURL: the base request url

Timeout: Duration until the client cut the connection (in milliseconds)

CacheElements: If it is <> 0, the connection pool will create an LRU Cache to store the response from calls of the indicated size

CacheStale: if true, will return expired elements in the cache is cant reach the destination.

Proxy: Proxy to use for each call.

MaxIdleConnsPerHost: Max idle Connections per host to use. If not specified, use the default (2)

###Questions?

Ask: 

fersca@hotmail.com
