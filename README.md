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
* make

This will generate gauge in the root directory

Installing
------------
````
make install
````

This installs gauge into __/usr/local__ by default. 
To install into a custom location use a prefix for installation

````
prefix=CUSTOM_PATH make install
````

Initializing a project
---------------------
In an empty directory initialize based on required language. Currently supported langauges are: Java, Ruby

````
gauge --init java
  or 
gauge --init ruby
````

Executing Specifications
---------------------
Inside the project directory

````
gauge specs/
  or
gauge specs/hello_world.spec
````

