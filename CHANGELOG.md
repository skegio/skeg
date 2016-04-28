# Changelog

## unreleased (TBD)

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
