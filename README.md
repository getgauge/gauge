Gauge
======

Gauge uses submodules. So issue the following commands before you attempt to make

```
  git submodule init
  git submodule update
```

List of submodules used


* [common](https://github.com/getgauge/common) - https://github.com/getgauge/common.git

Building
------------

````
go run make.go
````

This will generate gauge in the root directory

Installing
------------

````
go run make.go --install
````

This installs gauge into __/usr/local__ by default.
To install into a custom location use a prefix for installation

````
go run make.go --install --prefix CUSTOM_PATH
````

Initializing a project
---------------------
In an empty directory initialize a gauge project based on required language. Currently supported languages are: Java, Ruby

````
gauge --init java
````
For a gauge ruby project
````
gauge --init ruby
````

Executing Specifications
---------------------
Inside the project directory

To execute all specifications:
````
gauge specs/
````

To execute a single specification
````
gauge specs/hello_world.spec
````

