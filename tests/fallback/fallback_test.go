package cos_test

import (
	"fmt"
	"time"

	"github.com/rancher-sandbox/cOS/tests/sut"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("cOS booting fallback tests", func() {
	var s *sut.SUT

	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects()
	})
	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			s.GatherAllLogs()
		}
		if CurrentGinkgoTestDescription().Failed == false {
			s.Reset()
		}
	})

	Context("image is corrupted", func() {
		It("boots in fallback when rootfs is damaged, triggering a kernel panic", func() {
			currentVersion := s.GetOSRelease("VERSION")

			// Auto assessment was installed
			out, _ := s.Command("sudo cat /run/initramfs/cos-state/grubcustom")
			Expect(out).To(ContainSubstring("bootfile_loc"))

			out, _ = s.Command("sudo cat /run/initramfs/cos-state/grub_boot_assessment")
			Expect(out).To(ContainSubstring("boot_assessment_blk"))

			cmdline, _ := s.Command("sudo cat /proc/cmdline")
			Expect(cmdline).To(ContainSubstring("rd.emergency=reboot rd.shell=0 panic=5"))

			out, err := s.Command(fmt.Sprintf("elemental upgrade --no-verify --docker-image %s:cos-system-%s", s.GreenRepo, s.TestVersion))
			Expect(err).ToNot(HaveOccurred(), out)
			Expect(out).Should(ContainSubstring("Upgrade completed"))
			Expect(out).Should(ContainSubstring("Upgrading active partition"))

			out, _ = s.Command("sudo cat /run/initramfs/cos-state/boot_assessment")
			Expect(out).To(ContainSubstring("enable_boot_assessment=yes"))

			// Break the upgrade
			out, _ = s.Command("sudo mount -o rw,remount /run/initramfs/cos-state")
			fmt.Println(out)

			out, _ = s.Command("sudo mkdir -p /tmp/mnt/STATE")
			fmt.Println(out)

			s.Command("sudo mount /run/initramfs/cos-state/cOS/active.img /tmp/mnt/STATE")

			for _, d := range []string{"usr/lib/systemd"} {
				out, _ = s.Command("sudo rm -rfv /tmp/mnt/STATE/" + d)
			}

			out, _ = s.Command("sudo ls -liah /tmp/mnt/STATE/")
			fmt.Println(out)

			out, _ = s.Command("sudo umount /tmp/mnt/STATE")

			s.Reboot(700)

			v := s.GetOSRelease("VERSION")
			Expect(v).To(Equal(currentVersion))

			cmdline, _ = s.Command("sudo cat /proc/cmdline")
			Expect(cmdline).To(And(ContainSubstring("passive.img"), ContainSubstring("upgrade_failure")), cmdline)

			Eventually(func() string {
				out, _ := s.Command("sudo ls -liah /run/cos")
				return out
			}, 5*time.Minute, 10*time.Second).Should(ContainSubstring("upgrade_failure"))
		})
	})

	Context("GRUB cannot mount image", func() {
		When("COS_ACTIVE image was corrupted", func() {
			It("fallbacks by booting into passive", func() {
				Expect(s.BootFrom()).To(Equal(sut.Active))

				_, err := s.Command("mount -o rw,remount /run/initramfs/cos-state")
				Expect(err).ToNot(HaveOccurred())
				_, err = s.Command("rm -rf /run/initramfs/cos-state/cOS/active.img")
				Expect(err).ToNot(HaveOccurred())

				s.Reboot()

				Expect(s.BootFrom()).To(Equal(sut.Passive))
			})
		})
		When("COS_ACTIVE and COS_PASSIVE images are corrupted", func() {
			It("fallbacks by booting into recovery", func() {
				Expect(s.BootFrom()).To(Equal(sut.Active))

				_, err := s.Command("mount -o rw,remount /run/initramfs/cos-state")
				Expect(err).ToNot(HaveOccurred())
				_, err = s.Command("rm -rf /run/initramfs/cos-state/cOS/active.img")
				Expect(err).ToNot(HaveOccurred())
				_, err = s.Command("rm -rf /run/initramfs/cos-state/cOS/passive.img")
				Expect(err).ToNot(HaveOccurred())
				s.Reboot()

				Expect(s.BootFrom()).To(Equal(sut.Recovery))
			})
		})
	})
})
