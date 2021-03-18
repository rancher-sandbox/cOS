#!/bin/bash

# called by dracut
check() {
    return 255
}

# called by dracut
depends() {
    echo rootfs-block dm 
    return 0
}

# called by dracut
installkernel() {
    instmods overlay
}

# called by dracut
install() {
    declare moddir=${moddir}
    declare systemdutildir=${systemdutildir}
    declare systemdsystemunitdir=${systemdsystemunitdir}
    declare initdir="${initdir}"

    inst_multiple \
        mount mountpoint yip cos-setup sort findmnt rmdir
    inst_hook cmdline 30 "${moddir}/parse-cos-cmdline.sh"
    inst_script "${moddir}/cos-generator.sh" \
        "${systemdutildir}/system-generators/dracut-cos-generator"
    inst_script "${moddir}/cos-mount-layout.sh" "/sbin/cos-mount-layout"
    inst_script "${moddir}/cos-loop-img.sh" "/sbin/cos-loop-img"
    inst_simple "${moddir}/cos-immutable-rootfs.service" \
        "${systemdsystemunitdir}/cos-immutable-rootfs.service"
    inst_simple "${moddir}/cos-setup-rootfs.service" \
        "${systemdsystemunitdir}/cos-setup-rootfs.service"
    mkdir -p "${initdir}/${systemdsystemunitdir}/initrd-fs.target.requires"
    ln_r "../cos-immutable-rootfs.service" \
        "${systemdsystemunitdir}/initrd-fs.target.requires/cos-immutable-rootfs.service"
    ln_r "../cos-setup-rootfs.service" \
        "${systemdsystemunitdir}/initrd-fs.target.requires/cos-setup-rootfs.service"
    dracut_need_initqueue
}
