# Signed code

This is a working document where we try to take notes, collaborate on how to make a useful signed code mode that actually will increase safety. You are free to propose changes
or comment on the text below.

## What

We want to be able to sign Rye scripts and have the Rye evaluator that only runs scripts signed by specific keys.

## Questions

Where do we define public keys, so that attacker can't replace them?

How does the Rye runtime run, so that the attacker can't replace it or sideload it?

### Keys embedded in Rye binary

For Rye it's not uncommon for each project to have it's own binary with all the dependencies. So each project/app/app-server could also have it's own binary 
that has the keys it trusts embedded. In this way we join the two questions about replacing the keys and replacing the binary.

In this case we just have to make sure an attacker can't replace the binary with his own which he could easily build with different keys (Rye is open source).

Negative side of this approach is that we push trust (which keys will be added to the person that builds the Rye binary) which is not always the person that takes
care of the system and is trusted. And the trusted person must in every instance have the Go build and Rye dependencies installed. It would be better if trusted
person can just download and checkum a Rye binary from a trusted source.

### Keys on System keystore



### Keys in application directory

It's positive that all files are local.

Keys would need to have no write permissions (440) once set and Rye binary could check that and refuse to run otherwise. This would mean only a superuser could change them
once set.

Doubt: Couldn't the attacker just copy the Rye binary and scripts to a different folder where he sets up it's own pubkey file with 440 permissions?

## Scenarios

### Web-server


### Desktop app


### Mobile code
