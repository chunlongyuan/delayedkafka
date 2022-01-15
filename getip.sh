ETH_NAME=en0
LOCALIP=`ifconfig $ETH_NAME |sed -En 's/127.0.0.1//;s/.*inet (addr:)?(([0-9]*\.){3}[0-9]*).*/\2/p'`
