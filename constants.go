package main

// CONT_PREFIX is the prefix string used for containers, images, and volumes.
const CONT_PREFIX string = "skeg"

// DOCKER_HUB_ORG is the name of the organization on Docker Hub where
// predefined images can be found.
const DOCKER_HUB_ORG string = "skegio"

// ENVS_DIR is the directory in the user's homedir where data is created
const ENVS_DIR string = "skegs"

// IMAGE_VERSION defines image capabilities as defined by the BuildImage
// function.  When new capabilities are added to images, this number should be
// incremented to force building a new user image on environment creation.
//
//  version 0: user/tz creation
//  version 1: ssh key inclusion
const IMAGE_VERSION int = 1
