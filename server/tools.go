// +build tools

package main

// see: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
import (
	_ "github.com/Songmu/gocredits/cmd/gocredits"
	_ "github.com/golang/mock/gomock"
)
