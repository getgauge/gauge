Name:           gauge
Version:        <version>
Release:        <release>
Summary:        Cross-platform test automation
License:        GPLv3
URL:            http://getgauge.io/
Prefix:         /usr/local

Requires(post): /bin/sh

%description
Gauge is a light weight cross-platform test automation tool for authoring test cases in the business language.

%install
mkdir -p %{buildroot}/usr/local/bin/
cp %{_builddir}/bin/* %{buildroot}/usr/local/bin/
chmod +x %{buildroot}/usr/local/bin/*

%files
/usr/local/bin/gauge

%post
echo -e "\n\nWe are constantly looking to make Gauge better, and report usage statistics anonymously over time. If you do not want to participate please read instructions https://manpage.getgauge.io/gauge_telemetry_off.html on how to turn it off.\n"
    
%changelog
* Fri Apr 1 2016 ThoughtWorks Inc. <studios@thoughtworks.com>
- Release notes are available at https://github.com/getgauge/gauge/releases
