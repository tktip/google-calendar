
# google-calendar
  

**Introduction**

Google-calendar is used to communicate with the google APIs to retrieve, create, update, patch and delete google events. It supports running against multiple domains if thats needed.

  

**How to build**

  

1. Build as a runable file

  

We use make to build our projects. You can define what system to build for by configuring the GOOS environment variable.

  

  

>\> GOOS=windows make clean build

  

  

>\> GOOS=linux make clean build

  

  

These commands will build either a runable linux or windows file in the /bin/amd64 folder

  

  

2. Build a docker container

  

First you have to define the docker registry you are going to use in "envfile". Replace the REGISTRY variable with your selection.

  

Run the following command

  

  

>\> GOOS=linux make clean build container push

  

  

This will build the container and push it to your docker registry.

  

  

**How to run**

  

1. Executable

  

If you want to run it as a executable (Windows service etc) you will need to configure the correct environment variable. When you start the application set the CREDENTIALS environment to **file::\<location\>**

  

  

Windows example: **set "CREDENTIALS=Z://folder//cfg.json" & flyvo-calendar-queue.exe**

  

Linux example: **CREDENTIALS=/folder/cfg.json ./flyvo-calendar-queue**

  

  

2. Docker

  

You have to first mount the cfg file into the docker container, and then set the config variable to point to that location before running the service/container

**Configuration file**

The configuration file of google-calendar differs from the others. Its a json file using key/value to set up all the domains.

    {
      "domain1": {
        ...google credential content + scopes
      },
      "domain2": {
        ...google credential content + scopes
      }
    }
This will also change what APIs are exposed. All APIs starts with the domain path.
As an example if I only have one domain that i would like to name flyvo, the config file would look like this:

    {
      "flyvo": {
        "scopes": [
          "https://www.googleapis.com/auth/calendar",
          "https://www.googleapis.com/auth/calendar.events"
        ],
        "type": "",
        "project_id": "",
        "private_key_id": "",
        "private_key": "",
        "client_email": "",
        "client_id": "",
        "auth_uri": "https://accounts.google.com/o/oauth2/auth",
        "token_uri": "https://oauth2.googleapis.com/token",
        "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
        "client_x509_cert_url": ""
      }
    }

All our endpoints are exposed like this: /:domain/event/create
The :domain part is used as a regex where the user will write the domain it would like to use. In this case we would use http://localhost:8080/flyvo/event/create. The endpoint would then select the correct credentials with key flyvo from the example above. If you change the URL to http://localhost:8080/flyvo/event/create this will cause an invalid domain error, because we have no valid credentials for the domain "abc".