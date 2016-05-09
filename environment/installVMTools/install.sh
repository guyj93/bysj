echo "will install libvirt qemu-kvm and lxc"
read -r -p "continue?(enter 'y' to continue): " choice
if [ "$choice" = "y" ] ; then
	sudo apt-get install qemu-kvm lxc libvirt-bin libvirt-dev
fi
