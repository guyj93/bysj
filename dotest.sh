cd results

echo "-----local-----"
cd local
../../testTools/testScript.sh 10.0.0.1 2865 5201 6379 1234
cd ..

echo "-----lxcNetworkDefault-----"
cd lxcNetworkDefault
../../testTools/testScript.sh 10.0.0.1 12865 15201 16379 11234
cd ..

echo "-----lxcBridgeBr0-----"
cd lxcBridgeBr0
../../testTools/testScript.sh 10.0.0.2 2865 5201 6379 1234
cd ..

echo "-----kvmNetworkDefault-----"
cd kvmNetworkDefault
../../testTools/testScript.sh 10.0.0.1 22865 25201 26379 21234
cd ..

echo "-----kvmBridgeBr0-----"
cd kvmBridgeBr0
../../testTools/testScript.sh 10.0.0.3 2865 5201 6379 1234
cd ..
