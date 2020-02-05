import sys
import os
import json
import requests
import shutil
from subprocess import call


base_url = 'https://api.github.com/repos/getgauge/gauge/releases/latest'
try:
    latest_version = json.loads(requests.get(base_url).text)[0]['tag_name'].replace('v', '')
except KeyError as err:
    if os.getenv("GAUGE_VERSION") == None:
        message = requests.get(base_url).text
        raise Exception('Error while fetching the latest gauge version from github: {0}. Configure GAUGE_VERSION'.format(message))

def create_setup_file():
    tmpl = open("setup.tmpl", "r")
    setup = open("setup.py", "w+")
    version = os.getenv("GAUGE_VERSION") or latest_version
    name = os.getenv("GAUGE_PACKAGE_NAME") or "getgauge-cli"
    setup.write(tmpl.read().format(version,name))
    setup.close()
    tmpl.close()


def generate_package():
    shutil.rmtree('dist', True)
    print('Creating getgauge package.')
    create_setup_file()
    call([sys.executable, 'setup.py', 'sdist'], stdout=sys.stdout, stderr=sys.stderr)


def install():
    create_setup_file()
    call([sys.executable, 'setup.py', 'install'], stdout=sys.stdout, stderr=sys.stderr)

usage = """
Usage: python build.py --[option]

Options:
    --install :     installs getgauge-cli
    --dist    :     create pip package for getgauge-cli
"""


def main():
    if len(sys.argv) < 2:
        print(usage)
    else:
        if sys.argv[1] == '--install':
            install()
        elif sys.argv[1] == '--dist':
            generate_package()
        else:
            print(usage)

main()