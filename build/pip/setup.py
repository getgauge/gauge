"""The setup installs Gauge CLI through pip."""
from setuptools import setup
from os import path
from setuptools.command.install import install
import json
import platform
import glob
import zipfile
import shutil
import os
import sys
import requests

# TODO: bumped manually for now, wish to be automatically picked from source
release_ver = '1.0.2'
base_url = 'https://api.github.com/repos/getgauge/gauge/releases'

# Code to determine Architecture and OS where Gauge is being installed
arch_map = {"ia32": "x86", "x64": "x86_64"}
os_map = {"Darwin": "darwin", "linux": "linux", "win32": "windows"}
if '64' in platform.architecture()[0]:
    arch = arch_map['x64']
else:
    arch = arch_map['ia32']
os_name = os_map[platform.system()]
if os_name is 'win32':
    gauge_file = 'gauge.exe'
else:
    gauge_file = 'gauge'

latest_version = json.loads(requests.get(base_url).text)[0]['tag_name']


class CustomInstallCommand(install):
    """Customized setuptools install command to download and setup."""

    def _gauge_package_fetch(self):
        package_name = 'gauge-%s-%s.%s' % (release_ver, os_name, arch)
        package_url = base_url.replace('api.', '').replace('/repos', '') + '/download/v%s/%s.zip' % (release_ver, package_name)
        r2 = requests.get(package_url)
        with open("gauge.zip", "wb") as download:
            download.write(r2.content)

    def gauge_main_to_path(self):
        """Place Gauge CLI into Scripts Directory."""
        self._gauge_package_fetch()
        with zipfile.ZipFile('gauge.zip', "r") as z:
            z.extractall()
        target_path = os.path.join(self.install_scripts, gauge_file)
        source_path = os.path.join(os.getcwd(), gauge_file)
        if os.path.isfile(target_path):
            os.remove(target_path)
        if os.path.isfile(source_path):
            self.move_file(source_path, target_path)

            if os_name is not "win32":
                if sys.version_info[0] == 3:
                    os.chmod(target_path, 0o755)
                if sys.version_info[0] < 3:
                    os.chmod(target_path, 755)

    def run(self):
        """Custom Install / Run Commands."""
        self.gauge_main_to_path()
        install.run(self)


this_directory = path.abspath(path.dirname(__file__))
long_description = None
try:
    with open(path.join(this_directory, 'ReadMe.md'), 'rb') as f:
        long_description = f.read().decode('utf-8')
except IOError:
    long_description = '''Gauge is a free and open source test automation
    framework that takes the pain out of acceptance testing'''

setup(
    name='gauge-cli',
    version=release_ver,
    long_description=long_description,
    long_description_content_type='text/markdown',
    url='https://github.com/getgauge/gauge',
    platforms=["Windows", "Linux", "Unix", "Mac OS-X"],
    author='getgauge',
    author_email='revanth.mvs@hotmail.com',
    maintainer='Revant',
    license="GPL-3.0",
    cmdclass={
        'install': CustomInstallCommand,
    },
    classifiers=[
        "License :: GPL-3.0",
        "Development Status :: 5 - Production/Stable",
        "Topic :: Internet",
        "Topic :: Scientific/Engineering",
        "Topic :: Software Development",
        "Operating System :: Microsoft :: Windows",
        "Operating System :: Unix",
        "Operating System :: MacOS",
        "Programming Language :: Python",
        "Programming Language :: Python :: 2.7",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.4",
        "Programming Language :: Python :: 3.5",
        "Programming Language :: Python :: 3.6",
    ],
    install_requires=[
        'requests>=2.19.1',
    ],
)
print("\n***Gauge CLI - %s Installation Complete!***\n" % release_ver)
print("\n***Latest version of Gauge CLI is - %s***\n" % latest_version)
