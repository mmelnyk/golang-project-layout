# Import project parameters
include Project

include .make/setup.mk
include .make/help.mk

ifdef NEEDED_INIIALIZATION
	include .make/init.mk
else
	include .make/build.mk
	include .make/tools.mk
	ifeq ($(BUILD_MULTIPLATFORM),yes)
		include .make/buildmp.mk
	endif
	ifeq ($(BUILD_DOCKER),yes)
		include .make/docker.mk
	endif
endif
