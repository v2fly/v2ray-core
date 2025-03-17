
# OpenWRT

## build
```
git clone https://github.com/openwrt/openwrt -b openwrt-24.10
cd openwrt
./scripts/feeds update -a
./scripts/feeds install -a
make menuconfig
make -j1 V=s
```

## references
https://openwrt.org/docs/guide-user/installation/openwrt_x86
https://openwrt.org/ru/doc/howto/build?s[]=make&s[]=menuconfig
https://mirror-03.infra.openwrt.org/releases/24.10.0/targets/x86/64/
https://openwrt.org/zh/docs/guide-user/virtualization/qemu