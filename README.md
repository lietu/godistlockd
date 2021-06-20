# GoDistLock

Distributed lock manager.

*HEAVILY WORK IN PROGRESS, UNUSABLE IN CURRENT STATE*


## What is a lock manager?

Have you ever written a distributed system, such as a web application? Have you ever thought what happens when 2 things are at the same time trying to do the same thing?

For example on an e-commerce system you might want to keep track of stock. You might for example perform the following set of commands when someone places an order:

1. Check stock status
1. If item is out of stock return an error
1. Otherwise, deduct one from stock and create an order with that item

Now, imagine 2 requests coming in at almost exactly the same time, and they both perform the steps in sync. Both requests check stock status and it seems the product is out of stock, both deduct one from stock and create an order with that item. Now, at best you might start with 1 item in stock and end up with -1, at worst you lose track of the count and other DB integrity issues.

What do you do to solve this problem? You use locks. You add a step before checking stock status, to acquire an exclusive "lock" on the stock status for that product, and then after finishing the modifications you release that lock. This way only one request can at any point in time do things that depend on accurate information.
 
A lock manager is a tool that helps you manage the state of these "locks" and ensure that only one thing can hold an exclusive lock at a time.


## Why GoDistLock over other lock managers?

Other lock managers tend to have issues:

1. They depend on some other existing system you might not have any interest in running, e.g. Redis, MySQL, Memcached, etcd, or such.

1. Since they are built on some existing system they tend to not be built specifically with the needs of locking in mind, but rather with the limitations of that system in mind.

1. They often fail to ensure durability of the locks and can release a lock too early.

1. They rarely support more advanced uses of locks, such as "fencing", which can be used to further guarantee durability.

1. Many of the systems are not especially fault tolerant, and if a single server has issues the whole locking system fails.

1. They might have unexpected limits, such as that a client can only keep one lock at a time. Worst case, the previously held lock gets released without the developer noticing it.
 
GoDistLock has been built specifically with locking in mind, and while it likely is not perfect, it's aiming to be a step up from systems with above mentioned issues.

Having been built exclusively for locking using Golang, GoDistLock should be very lightweight and achieve performance that exceeds that of similar systems built on more complex platforms.

GoDistLock also is built from ground up to help developers ensure durability, for example you will not be given a lock if the cluster of servers is unable to get a majority agreement that you can get exclusive access to the lock.

Additionally the fence token system will allow developers to build services that can check that whoever is making changes is actually the one with the latest token, as locks may get released due to timeouts for various reasons.


## Clustering

As mentioned above, GoDistLock supports connecting multiple servers to a cluster. In a cluster the servers will automatically cooperate to try and ensure durability of the locks as they best can.

When a lock is requested by a client, the servers check all other servers in the cluster if they think it's ok to give the lock, and only give the lock if they're able to save the new status to a majority of the cluster's servers. In short, acquiring a lock requires >50% quorum.
  
This also ensures fault tolerance as long as the clients know to switch to another server if connections to one fail, as the cluster does not require 100% of the servers to be available, just the majority. 

For example if you are running a cluster of 3 servers, 1 server can go down and the rest can continue operation. With 5 servers, 2 servers can go down, and so on as long as >50% of servers are operational.

Sharding is left to the user due to the vast number of possible sharding strategies users might need.


## Known issues

If a server dies while clients are holding locks, they cannot release them anymore. Would be nice if a client could reconnect to another server and release the locks? Probably shouldn't release any locks the server was holding when connection to it dies in the cluster?

When a server connects to the relay network it should likely synchronize the current status of locks from the other servers so you can do a rolling update between servers.

A server should react to a "gentle" SIGINT/SIGTERM by preventing incoming lock requests, and blocking until it's locks have been released or time out.


## Ideas, research, etc.


https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html


# Financial support

This project has been made possible thanks to [Cocreators](https://cocreators.ee) and [Lietu](https://lietu.net). You can help us continue our open source work by supporting us on [Buy me a coffee](https://www.buymeacoffee.com/cocreators).

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/cocreators)
