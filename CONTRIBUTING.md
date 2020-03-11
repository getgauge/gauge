## Contributing to Gauge

Contributions to Gauge are welcome and appreciated. Coding is definitely not the only way you can contribute to Gauge. There are many valuable ways to contribute to the product and to help the growing Gauge community.

Please read this document to understand the process for contributing.

## Different ways to contribute

* You can [report an issue](https://github.com/getgauge/gauge/issues) you found
* Help us [test Gauge](https://github.com/getgauge/gauge-tests) by adding to our existing automated tests
* Help someone get started with Gauge on [our discussion forum](https://groups.google.com/forum/#!forum/getgauge)
* Contribute [to our blog](https://gauge.org/blog/) 
* Add to our [set of examples](https://docs.gauge.org/examples.html) to help someone new to Gauge get started easily
* Help us improve [our documentation](https://github.com/getgauge/documentation)
* Contribute code to Gauge! 

All repositories are hosted on GitHub. Gauge’s core is written in Golang but plugins are, and can be, written in any popular language. Pick up any pending feature or bug, big or small, then send us a pull request. Even fixing broken links is a big, big help!

## How do I start contributing

There are issues of varying levels across all Gauge repositories. All issues that need to be addressed are tagged as _'Help Needed'_. One easy way to get started is to pick a small bug to fix. These have been tagged as _'Easy Picks'_.

If you need help in getting started with contribution, feel free to reach out on the [Google Groups](https://groups.google.com/forum/#!forum/getgauge) or [Spectrum](https://spectrum.chat/gauge).

### Gauge Core v/s Plugins

Gauge Core is a project that has features that would reflect across all Gauge use cases. These features are typically agnostic of the user's choice of implementation language.

Plugins are meant to do something specific. These could be adding support for a new language, or have a new report etc. So, depending on where you see your contribution fit, please focus on the respective repository.

If your contribution is a code contribution and you do send us a pull request, you will first need to read and sign the [Contributor License Agreement](https://gauge-bot.herokuapp.com/cla/).

### Developer documentation

If you are trying to write plugins for Gauge or trying to contribute to Gauge core, take a look at the [Developer Documentation](https://github.com/getgauge/gauge/wiki/Gauge-Technical-Documentation).


## Bump up gauge version

* Update the value of `CurrentGaugeVersion` variable in `version/version.go` file.

Ex:
```diff
 // CurrentGaugeVersion represents the current version of Gauge
-var CurrentGaugeVersion = &Version{1, 0, 7}
+var CurrentGaugeVersion = &Version{1, 0, 8}

```