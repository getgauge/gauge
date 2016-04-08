## Generating .deb

### Requirements

- Debian 7 or Ubuntu 14.04 or Fedora 22+
- Install these packages:
    - For Ubuntu/Debian:

      ```
      $ sudo apt-get install build-essential fakeroot
      ```

    - For Fedora/RHEL/CentOS:

      ```
      $ sudo dnf install make automake gcc gcc-c++ kernel-devel dpkg-dev fakeroot
      ```

### Run the script

Run this command from the root of this project:

```
$ ./build/mkdeb.sh
```

## Generating .rpm

### Requirements

- RHEL6+, CentOS 7+ or Fedora 22+
- Install these packages:

    ```
    $ sudo yum install make automake gcc gcc-c++ kernel-devel rpmdevtools
    ```

### Run the script

Run this command from the root of this project:

```
$ ./build/mkrpm.sh
```
