# Port mapping hooks
We can use these hooks to do port mapping for VMs in NAT network.  
(1) Check and correct the IPs and ports in the hooks  
(2) Copy these hooks into /etc/libvirt/hooks  
(3) Restart the libvirtd process by run
```
/etc/init.d/libvirt-bin restart
```
or just roughly kill the libvirtd process and normally it will restart.

# Use Open vSwitch with libvirt on Ubuntu
Profile of apparmor didn't allow libvirtd to excute programs in /usr/local/bin, which is the default directory contains the "ovs-vsctl".

So we must modify the profile mannually to let libvirt work with ovs.

(1) open the apparmor profile of libvirtd
```
sudo gedit /etc/apparmor.d/usr.sbin.libvirtd
```
(2) find lines of:
```
  /bin/* PUx,
  /sbin/* PUx,
  /usr/bin/* PUx,
  /usr/sbin/* PUx,
```
(3) insert these two lines (notice ovs-vsctl is contained in /usr/local/bin):
```
  /usr/local/bin/* PUx,
  /usr/local/sbin/* PUx,
```
(4) save and exit gedit

(5) reload the profile with the command:
```
sudo /etc/init.d/apparmor reload
```
Now, we can successfully start VMs connected to an ovs bridge!

