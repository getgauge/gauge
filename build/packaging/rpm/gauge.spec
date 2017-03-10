Name:           gauge
Version:        <version>
Release:        <release>
Summary:        Cross-platform test automation
License:        GPLv3
URL:            http://getgauge.io/
Prefix:         /usr/local

Provides: gauge_screenshot

Requires(post): /bin/sh

%description
Gauge is a light weight cross-platform test automation tool.
It provides the ability to author test cases in the
business language.

%install
mkdir -p %{buildroot}/usr/local/gauge/config
cp -r %{_builddir}/config/* %{buildroot}/usr/local/gauge/config
mkdir -p %{buildroot}/usr/local/bin/
cp %{_builddir}/bin/* %{buildroot}/usr/local/bin/
chmod +x %{buildroot}/usr/local/bin/*

%files
/usr/local/bin/gauge
/usr/local/bin/gauge_screenshot
/usr/local/bin/gauge_setup
/usr/local/gauge/config/gauge.properties
/usr/local/gauge/config/notice.md
/usr/local/gauge/config/skel/env/default.properties
/usr/local/gauge/config/skel/example.spec

%post
echo -e "\n\nPlease run the 'gauge_setup' command to complete installation.\n"

%changelog
* Fri Apr 1 2016 ThoughtWorks Inc. <studios@thoughtworks.com>
- Release notes are available at https://github.com/getgauge/gauge/releases
