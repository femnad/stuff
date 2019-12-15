# stuff #

A container repo for throwaway stuff.

## Role call ##

### abry ###

Lazily add [fish shell](https://github.com/fish-shell/fish-shell) abbreviations. Hardcodes the abbreviations file like a boss.

### clrf ###

Raise on the shoulders of [Clipster](https://github.com/mrichar1/clipster) and [Rofi](https://github.com/DaveDavenport/rofi) to copying strings between primary and clipboard selections.

### einy ###

Add a host name to a group in [Ansible inventory file](http://docs.ansible.com/ansible/latest/intro_inventory.html#hosts-and-groups). Don't know if it'll work with default group, or at all.

### fred ###

A [rofi](https://github.com/DaveDavenport/rofi/) helper for [pass](https://www.passwordstore.org/) interaction.

### hazy ###

Add a hostname for a host to the user's SSH configuration file. Tries really hard not to mess up with the existing file. But is it enough?

### klen ###

Convert multiples of bytes to other multiples of bytes.

### pand ###

Add a string to [fish-shell](https://github.com/fish-shell/fish-shell) history file, useful for a command has been `eval`ed but it'd be useful to have in history, too.

### rabn ###

List the directories under a given path if run with no arguments, else add the argument to a history file and print it to standard output. Intended to be used with [fzf](https://github.com/junegunn/fzf). Example for fish shell:

```fish
rabn -path ~/repos -history-file ~/.rabn_repos | fzf +s | read selection; rabn -history-file ~/.rabn_repos $selection
```

### wstr ###

Append a profile to AWS credentials file, but like a cave person, i.e. do not check if the profile already exists. Was useful for testing [Minio](https://github.com/minio/minio) once.

### qrdm

Calculate the SHA256 checksum from the contents of the target URL.
