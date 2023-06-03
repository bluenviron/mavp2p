include $(TOPDIR)/rules.mk

PKG_NAME:=mavp2p
PKG_VERSION:=v0.0.0
PKG_RELEASE:=1

PKG_SOURCE_PROTO:=git
PKG_SOURCE_URL:=https://github.com/bluenviron/mavp2p
PKG_SOURCE_VERSION:=$(PKG_VERSION)

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

GO_PKG:=github.com/bluenviron/mavp2p
GO_PKG_LDFLAGS_X:=github.com/bluenviron/mavp2p/internal/core.version=$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

GO_MOD_ARGS:=-buildvcs=false

define Package/mavp2p
  SECTION:=net
  CATEGORY:=Network
  TITLE:=mavp2p
  URL:=https://github.com/bluenviron/mavp2p
  DEPENDS:=$(GO_ARCH_DEPENDS)
endef

define Package/mavp2p/description
  flexible and efficient Mavlink router
endef

$(eval $(call GoBinPackage,mavp2p))
$(eval $(call BuildPackage,mavp2p))
