![Gauge Logo](Gauge-Logo.png)


[![Actions Status](https://github.com/getgauge/gauge/workflows/build/badge.svg)](https://github.com/getgauge/gauge/actions)

[![Ask at StackOverflow](https://img.shields.io/badge/StackOverflow-getgauge-F5C10E.svg?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAAnFBMVEUAAADs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PHs8PG3iEVjAAAAM3RSTlMAAQMGBwkLDA0QExweKS4wMzc4PkdJT1JdY2dvc3WDjpiam6avtcfMz9ng4ubp7fHz9/tqGqSaAAAAfUlEQVQYGV3BBw6CQABFwb/YUOy9IIpgb8i7/92ErBLCjKzGKzAqa185tVTmhCR9/ZmoJy1gY2QN4DIy3oNjXVY3gvu0HfN09dPZwXsdRsqNb8GkY5rblE9NuRWZNF7O9r4sx5sfEjKeJJDlDv2zAwJREAhEQSBQCYgKUfEFJ7oYF2usUEAAAAAASUVORK5CYII=)](https://stackoverflow.com/questions/ask?tags=getgauge)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v1.4%20adopted-ff69b4.svg)](CODE_OF_CONDUCT.md)


## Important note

Gauge's official sponsorship ended in 2021. Despite this, [we](https://github.com/getgauge/gauge/graphs/contributors) strived to maintain the project with our personal resources as best as possible. However, we've reached a point where we can't ensure the same level of dedication.

The core team will only look into issues in our spare time, which means slower response times. We welcome pull requests, but please anticipate significant delays in reviewing and merging.

Moreover, please consider using Gauge only if you're prepared to adopt and support it fully on your own. With our shift to community-driven maintenance, primary responsibility will fall onto individual users and contributors.

We appreciate your understanding, contributions, and continued support.

# Welcome to Gauge

Gauge is a light weight cross-platform test automation tool. It provides the ability to author test cases in the business language.

## Get Started
Read more about [Why Gauge](https://gauge.org/2018/05/15/why-we-built-gauge/) can be used, its [terminologies](https://docs.gauge.org/writing-specifications.html) and [**get started...**](https://docs.gauge.org/getting_started/installing-gauge.html)

## Find out more

| **[User Docs][userdocs]**     | **[Setup Guide][get-started]**     | **[Examples][examples]**           | **[Contributing][contributing]**           | **[Brand Style Guide][branding]**           |
|:-------------------------------------:|:-------------------------------:|:-----------------------------------:|:---------------------------------------------:|:---------------------------------------------:|
| [![i1][userdocs-image]][userdocs]<br>Learn more about using Gauge | [![i2][getstarted-image]][get-started]<br> Getting started with Gauge | [![i3][examples-image]][examples]<br>Some Gauge Examples | [![i4][contributing-image]][contributing]<br>How can you contribute to Gauge? | [![i5][branding-image]][branding]<br>Gauge brand colours, logos, images etc. |

[userdocs-image]:https://d3i6fms1cm1j0i.cloudfront.net/github/images/techdocs.png
[getstarted-image]:https://d3i6fms1cm1j0i.cloudfront.net/github/images/setup.png
[examples-image]:https://d3i6fms1cm1j0i.cloudfront.net/github/images/roadmap.png
[contributing-image]:https://d3i6fms1cm1j0i.cloudfront.net/github/images/contributing.png
[branding-image]:https://d3i6fms1cm1j0i.cloudfront.net/github/images/setup.png

[userdocs]:https://docs.gauge.org/
[get-started]:https://github.com/getgauge/gauge/wiki/Setup
[examples]:https://getgauge-examples.github.io/
[contributing]:CONTRIBUTING.md
[branding]:https://brand.gauge.org/

## Authenticating plugin downloads

`gauge install <plugin>` fetches plugin releases from GitHub. Unauthenticated
GitHub traffic is rate limited hard enough that installs fail on shared CI
runners — often as a `504 Gateway Timeout` while GitHub sheds load rather than
an explicit rate-limit error, which is why the failure does not look like one.

Set a token and Gauge sends it as `Authorization: Bearer <token>`:

| Variable | Notes |
|---|---|
| `GAUGE_GITHUB_TOKEN` | Checked first. Use it to give Gauge a token without widening what every other tool in the job sees. |
| `GITHUB_TOKEN` | Used when `GAUGE_GITHUB_TOKEN` is unset. On GitHub Actions this is the automatic per-job token. |

A read-only token is enough — plugin releases are public, and the token is only
there to raise the rate limit.

```yaml
- run: gauge install java
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

The token is sent only to `github.com`, `www.github.com`, `api.github.com` and
`codeload.github.com`. It is never sent to a redirect target such as
`objects.githubusercontent.com`, which is where release assets actually live
and which rejects requests that carry an `Authorization` header. It is never
logged: it is not part of any URL, and nothing prints request headers.

## Questions or need help?

### Troubleshooting
Context specific **Troubleshooting** guide is available in relevant pages of the [Gauge Documentation](https://docs.gauge.org/).

### Talk to us

Please see below for the best place to ask a query:

- How do I? -- [Discussions](https://github.com/getgauge/gauge/discussions)
- I got this error, why? -- [Discussions](https://github.com/getgauge/gauge/discussions)
- I got this error and I'm sure it's a bug -- file an [issue](https://github.com/getgauge/gauge/issues)
- I have an idea/request -- file an [issue](https://github.com/getgauge/gauge/issues)
- Why do you? -- [Discussions](https://github.com/getgauge/gauge/discussions)
- When will you? -- [Discussions](https://github.com/getgauge/gauge/discussions)

### Other projects

Also maintained by Gauge

- [Taiko](https://github.com/getgauge/taiko) Headless web browser automation.

## Add Gauge Badge to your project!
Copy the text below and add it to the README.md of your project using Gauge, and show your support. It looks like this:

[![Gauge Badge](https://gauge.org/Gauge_Badge.svg)](https://gauge.org)

```
[![Gauge Badge](https://gauge.org/Gauge_Badge.svg)](https://gauge.org)
```

## License

Gauge is released under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.

## Copyright

Copyright 2018 ThoughtWorks, Inc.
