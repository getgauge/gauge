Name:           gauge
Version:        <version>
Release:        1%{?dist}
Summary:        Cross-platform test automation
License:        GPLv3
URL:            http://getgauge.io/
Prefix:         /usr

Provides: gauge_screenshot

Requires(post): /bin/sh

%description
Gauge is a light weight cross-platform test automation tool.
It provides the ability to author test cases in the
business language.

%install
mkdir -p %{buildroot}/usr/share/gauge/
cp -r %{_builddir}/share/gauge/* %{buildroot}/usr/share/gauge/
mkdir -p %{buildroot}/usr/bin/
cp %{_builddir}/bin/* %{buildroot}/usr/bin/
chmod +x %{buildroot}/usr/bin/*

%files
/usr/bin/gauge
/usr/bin/gauge_screenshot
/usr/bin/gauge_setup
/usr/share/gauge/gauge.properties
/usr/share/gauge/notice.md
/usr/share/gauge/skel/env/default.properties
/usr/share/gauge/skel/example.spec

%post
echo -e "\n\nPlease run the 'gauge_setup' command to complete installation.\n"

%changelog
* Fri Apr 1 2016 ThoughtWorks Inc. <studios@thoughtworks.com>
- Release notes are available at https://github.com/getgauge/gauge/releases
