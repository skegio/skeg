# Changelog

## unreleased (TBD)

## v0.4.0 (2018-01-26)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.4.0)

* add `ssh-config` command to print out ssh config for an env
* handle ssh key inclusion with more flexibility
* limit user image search to the version that skeg is using
* add `ssh` alias for `connect`
* add `ls` alias for `list`

## v0.3.0 (2017-02-28)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.3.0)

* add --volume-home option to use a Docker volume for homedir instead of host mount
* move ssh key into user image, so less set up required for homedir

## v0.2.4 (2016-12-26)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.2.4)

* add default windows docker endpoint

## v0.2.3 (2016-12-23)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.2.3)

* update vendor dependencies
* add windows build

## v0.2.2 (2016-12-04)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.2.2)

- allow specifying a different image when running `skeg rebuild ...`
- updated core images to include Go 1.7 and 1.8 (beta)

## v0.2.1 (2016-05-23)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.2.1)

* Ensure SSH port is functional before connecting

## v0.2 (2016-05-02)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.2)

* Detect timezone, if possible
* Add timezone to image list
* Show environment count in image list, to aid in cleanup
* When creating new environments or rebuilding existing ones, use the previous timezone, if set.
* Add `vendor/` code, needs Go 1.5/1.6 to build now
* Fix environment destruction when adding volume in environment homedir

## v0.1.4 (2016-04-22)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.1.4)

* **Breaking change**: Containers now contain the username, for multi-user systems.  When `skeg` detects this, it will display instructions on how to fix the issue.

## v0.1.3 (2016-04-15)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.1.3)

* Properly handle ssh port bound to non 0.0.0.0 ip (to support [dlite](https://github.com/nlf/dlite))

## v0.1.2 (2016-03-27)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.1.2)

* Allow creating environment when container is missing but environment directory exists
* Protect against duplicate environments when using the `skeg run` command

## v0.1.1 (2016-03-27)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.1.1)

* This changes the directory that local environments are created in from $HOME/envs to $HOME/skegs.

## v0.1 (2016-03-25)

[Downloads](https://github.com/skegio/skeg/releases/tag/v0.1)

Initial release.

* Base functionality exists (create, connect, rebuild, destroy).
