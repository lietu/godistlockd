# The godistlock communications protocols explained

Glossary of identifiers:

 - `<id>`: The server's ID, could e.g. be FQDN, or whatever the admin has set
 - `<nonce>`: A unique identifier to connect response to this query, i.e. message `IS foo 123` might get a response `NO 123`
 - `<version>`: Version identifier
 - `<fence>`: A nonce for the holding of the lock, every time the lock is acquired a new unique fence is generated
 - `<lock>`: A unique name for a lock
 
Since the protocol is a text protocol, none of the identifiers may contain spaces.
Each message consists of the keyword (e.g. `HELLO`), space separated arguments, and a newline (always `\n`).


## Client protocol

### Messages client -> server

 - `HELLO <version> <nonce>` -> Hi, I'm a client running version <version>
 - `ON <lock> <timeout> <nonce>` -> Wait until you get lock, keep locked until timeout, will return a token for fencing
 - `OFF <lock> <nonce>` -> Release lock
 - `TRY <lock> <timeout> <nonce>` -> Check if you can get lock, get it if you can, will return a token for fencing
 - `REFRESH <lock> <fence> <timeout> <nonce>` -> I want to keep this lock for a bit longer
 - `IS <lock> <nonce>` -> Check if the lock is engaged, returns fence token (nonce) if it is
 - `STATS <nonce>` -> Get count of locks and other stats about the system

### Responses server -> client

 - `HELLO <nonce> <id> <version>` -> Hi, I'm <id> running <version>
 - `GIVE <nonce> <fence>` -> Here you go, you now have the lock
 - `LOCK <nonce> <fence>` -> Yes, lock <lock> is locked, this is the <fence> token
 - `NO <nonce>` -> Lock <lock> is not locked
 - `STATS <nonce> <name> <value>` -> Stats response
 - `STATSEND <nonce>` -> All stats responses have been sent
 - `ERR <nonce> <msg>` -> System error, you will be disconnected, maybe try another server


## Relay protocol server <-> server

### Commands / requests

 - `HELLO <id> <version> <nonce>` -> I'm server <id> running <version>
 - `PROP <lock> <nonce>` -> I propose locking, please give me your lock status
 - `SCHED <lock> <nonce>` -> We have quorum, nobody is locked, prep to lock
 - `COMM <lock> <timeout> <nonce>` -> Commit lock with X timeout
 - `OFF <lock> <nonce>` -> Release lock if it was held by the source relay

### Responses

 - `HOWDY <nonce> <id> <version>` -> Hi, I'm <id> running <version>
 - `STAT <nonce> <status>` -> Response to PROP: status 0 = ok, 1 = held by this server, 2 = held by another relay
 - `ACK <nonce> <ok>` -> Acknowledging SCHED: ok 1 = ok, 0 = err
 - `CONF <nonce>` -> Confirming commit 1/0 = ok/err
