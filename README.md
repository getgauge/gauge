Gauge
======

Gauge uses submodules. So issue the following commands before you attempt to make

```
  git submodule init
  git submodule update
```

Building
-----------

````
go run build/make.go
````

This will generate gauge in the root directory

Running Tests
-------------

````
go test
````
or 
````
go run build/make.go --test
````
With Test coverage
````
go run build/make.go --test --coverage
````

Installing
------------

###MacOS and Linux

````
go run build/make.go --install
````

This installs gauge into __/usr/local__ by default.
To install into a custom location use a prefix for installation

````
go run build/make.go --install --prefix CUSTOM_PATH
````

###Windows

````
go run build\make.go --install --prefix CUSTOM_PATH
````

Set environment variable GAUGE_ROOT to the CUSTOM_PATH

Initializing a project
---------------------
In an empty directory initialize a gauge project based on required language.

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


