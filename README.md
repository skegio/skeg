# skeg

> Simple Docker development containers.

See <http://skeg.io> for usage and install instructions.

# Future possibilities

* Non-Ubuntu base images
* Mount the Docker socket
* Mount certain paths in every dev container
    * dotfiles (vim/emacs/git/etc.)
    * shared secrets or configuration
* Integrate with docker-compose

# Thank you

This project wouldn't be possible without the following libraries:

* From the Docker project
    * [jsonmessage](https://github.com/docker/docker/pkg/jsonmessage)
    * [go-connections](https://github.com/docker/go-connections/nat)
* [go-dockerclient](https://github.com/fsouza/go-dockerclient)
* [go-flags](https://github.com/jessevdk/go-flags)
* [logrus](https://github.com/Sirupsen/logrus)
* [testify](https://github.com/stretchr/testify)

# License

Copyright © 2016 Nate Jones

Distributed under the Apache License Version 2.0.
