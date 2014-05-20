Welcome to Gauge
================

This is an executable specification file. This file follows markdown syntax. Every heading in this file denotes a scenario. Every bulleted point denotes a step.

To execute this specification, use

	gauge spec/hello_world.spec

* A context step which gets executed before every scenario

Say hello to Gauge
------------------

tags: hello world, first test

* Say "hello" to "gauge"


Getting started with Gauge
---------------------------

This is the second scenario in this specification

* Say "hello again" to "gauge"
* Step that takes a table
    |product|description                 |
    |-------|----------------------------|
    |Gocd   |Continous delivery platform |
    |Twist  |BDD style automation testing|
    |Gauge  |Next generation of Twist    |

