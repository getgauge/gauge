Specification Heading
=====================

This is an executable specification file. This file follows markdown syntax. Every heading in this file denotes a scenario. Every bulleted point denotes a step.

To execute this specification, use

	gauge spec/hello_world.spec

* A context step which gets executed before every scenario

First scenario
--------------

tags: hello world, first test

* Say "hello" to "gauge"


Second scenario for the specification
-------------------------------------

This is the second scenario in this specification

* Say "hello again" to "gauge"
* Step that takes a table
    |Product|       Description           |
    |-------|-----------------------------|
    |Gauge  |Test automation with ease    |
    |Mingle |Agile project management     |
    |Snap   |Hosted continuous integration|
    |Gocd   |Continuous delivery platform |

