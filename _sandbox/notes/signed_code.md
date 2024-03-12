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

Negative is that this is OS dependant. Furthermore, it seems that even specific Linux distributtions don't have a unified keystore. 
Concretely it seems Ubuntu doesn't have a universal "system keystore"; it depends on the specific setup. So it's again not something immutable.

TODO: Look at examples of  https://github.com/miekg/pkcs11 and github.com/ThalesIgnite/crypto11 . But it seems this is generally not that recommended way citing
complexity, portability and security (such program needs additional permissions on the system).

### Keys in application directory

It's positive that all files are local.

Keys would need to have no write permissions (440) once set and Rye binary could check that and refuse to run otherwise. This would mean only a superuser could change them
once set.

Doubt #1: Couldn't the attacker just copy the Rye binary and scripts to a different folder where he sets up it's own pubkey file with 440 permissions?

### A specific Linux user-level path

If we say that a Linux user is the unit of trust. Then a program running as some user could look for user-specific path. If you copy the structure (#1) within the same user and 
this file has 440 permissions then you still can't replace the pubkey files. In this case an attacker would maybe have to setup another user to copy the app. Generally creating another 
user requires superuser access.

### Signed Rye binary

Signed Rye binary would prevent attacker from building it's own binary with proper signing certificate. So attacker couldn't change the embedded public keys if we used that. Or the
path to pubkey file.

Doubt: could attacker still copy the file and inject his own pubkey?
Doubt: could attacker build and run his own unsigned Rye?

### Apparmor

Apparmor could limit (whitelist) programs that can be started under one folder / user. This would prevent running a unsingned or attacker built copy in anohter location. 

````
# Define the user and folder path
  unconfined // This profile applies to all unconfined domains (processes)

  # Restrict execution within the user's folder (replace 'username' and 'folder_path')
  deny /home/username/folder_path/* (executable)

  # Allow execution of the specific program (replace 'allowed_program')
  allow /home/username/folder_path/allowed_program (execute)

````

Apparmour can't enforce only signed apps be ran, but it can leverage a third signature checking program. Gpg or pcks11 are tools that could be used for checking signatures.

````
# Allow execution only if signature is valid
  deny /path/to/your/program ( if ( program_signature_is_valid() != 1 )
````

TODO: explore this further

## Scenarios

### Web-server


### Desktop app


### Mobile code
