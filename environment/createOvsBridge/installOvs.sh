# exit when use unset variable
set -u;
# exit after an error command
set -e;

#requirement
sudo apt-get install git fuse libfuse-dev dh-autoreconf openssl libssl-dev

#unzip
tar zxvf openvswitch-2.5.0.tar.gz

#install
cd openvswitch-2.5.0
./boot.sh
./configure --with-linux=/lib/modules/`uname -r`/build
make
sudo make install
sudo make modules_install
sudo modprobe openvswitch
sudo mkdir -p /usr/local/etc/openvswitch
sudo ovsdb-tool create /usr/local/etc/openvswitch/conf.db vswitchd/vswitch.ovsschema
cd ..

#startup
./startupOvs.sh
