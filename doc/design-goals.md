# Design Goals

## From a Users Perspective

## Internal to code

Plugins (middleware) are attached to the code by simply copying a file with the entire plugin into
the pulugin directory.   The file has an "init" function that calls and saves its configuration with
the server.   This means that all of the different types of plugins will need to have a common
set of interfaces.


