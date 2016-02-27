## Generating .deb

### Requirements

- Debian 7 or Ubuntu 14.04 or Fedora 22+
- Install these packages:
    - For Ubuntu/Debian: `$ sudo apt-get install build-essential fakeroot`
    - For Fedora: `$ sudo dnf install make automake gcc gcc-c++ kernel-devel dpkg-dev`

 ### Run the script

Run this command from the root of this project:

 ```
 $ ./build/mkdeb.sh
 ```
