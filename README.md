# Brainf*ck-as-a-Service

A little BF interpreter, inspired by modern systems design trends.

## How to run it?

```
docker-compose up -d
bash hello.sh # Should print "Hello World!"
```

## How does it work?

Microservices! There is some nginx gateway, that does load balancing to the API instances.
Each API instance (see `./api` directory) handles parsing and loops.
Of course, nobody in his sane mind would implement a stack manually these days, so we use Redis for stack operations.
Pointer movements are controlled by another microservice, `./ptr`. This one uses MongoDB, because, well, what could possibly go wrong?
Finally, there is memory access service (`./mem`) that uses Postgres as a memory storage, battle-tested, scalable technology.

As a result, a typical hello world script runs in ~400ms on a modern MacBook, not bad at all!

80% of the code was copied from StackOverflow, so it should be bug-free. Enjoy!
