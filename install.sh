
os_name=$(uname -s | tr '[:upper:]' '[:lower:]')
os_arch=$(uname -m)
latest_release='0.3.0'
file_name=gauge-$latest_release'-'$os_name.$os_arch
base_url='https://github.com/getgauge/gauge/releases/download/v'$latest_release'/'$file_name

install_linux_image() {
	wget $base_url'.zip'
	unzip $file_name'.zip' -d $file_name
	cd $file_name
	./install.sh $@
	cd ..
	rm -rf $file_name $file_name'.zip'
}

install_darwin_image() {
	if ! type brew > /dev/null; then
		echo 'ERROR: HomeBrew is not installed on this system.'
		echo '\nPlease download the installer from '$base_url'.pkg and install manually.'
		echo 'You will have to install the language and IDE plugins seperately.'
		exit 1
	else
		brew install gauge
	fi
}

if [ $os_name == 'darwin' ]; then
	install_darwin_image $@
elif [ $os_name == 'linux' ]; then
	install_linux_image $@
fi

for arg in "$@"; do
	gauge --install $arg
done
