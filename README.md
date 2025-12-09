# Go Load Balancer

This is a load balancer written in Go. I'm trying to go whole-hog on this project (As Kent Beck says in _Small Talk Best Practice Patterns_). So I'm using the book Learning Go by Jon Bodner and the book 100 Go Mistakes and How To Avoid Them by Teiva Harsanyi. My goal for this project is to learn a lot more about network programming, data structures & algorithms, and writing production-level Go applications. 

## How does a load balancer work?

A load balancer is just a reverse proxy that chooses which server to forward a request to. And a reverse proxy is just a server that sits in front of another server acting as the pass through. So if I have a website that I'm serving to people at 0.0.0.0/24:8000. I can have a reverse proxy at some other IP address 0.0.0.0/24:443 and that's where I'll point my dns resolution. So it would work something like this

User sends request:

User -> MyCoolWebsite.com -> 0.0.0.0/24:443 (Reverse Proxy) -> 0.0.0.0/24:8000

Server sends response:
0.0.0.0/24:8000 -> 0.0.0.0/24:443 -> User

So. The Reverse Proxy in this case will be a load balancer and would act something like this:

                             --------------------------
                             +                        + -> 0.0.0.0/0:8000
                             +                        +
                             +                        +
                             +                        +
User -> MyCoolWebsite.com -> + 0.0.0.0/24:443         + -> 0.0.0.0/4:8000
                             +  Load balancer chooses +
                             +  which server to       +
                             +  forward the request   +
                             +  to                    + 
                             +                        + -> 0.0.0.0/8:8000
                             +                        +
                             +                        +
                             +                        +
                             +                        + -> 0.0.0.0/12:8000
                             --------------------------

## What is the scope of this project?

The goal here is to learn more about the standard library and to understand more about networking. I will NOT use any of the built in HTTP utilities I will be interpreting RAW requests and handling TCP actions. 

