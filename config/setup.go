package config

import (
	"r2/config/route"
)

func Clear() error {
	err := route.ClearWan(V.Wan)
	if err != nil {
		return err
	}
	err = route.ClearLan(V.Lan)
	if err != nil {
		return err
	}
	err = route.ClearLo(V.Lo)
	if err != nil {
		return err
	}
	err = route.ClearTProxy()
	if err != nil {
		return err
	}
	return nil
}

func Setup() error {
	err := route.SetupTProxy()
	if err != nil {
		return err
	}
	err = route.SetupWan(V.Wan)
	if err != nil {
		return err
	}
	err = route.SetupLan(V.Lan)
	if err != nil {
		return err
	}
	err = route.SetupLo(V.Lo)
	if err != nil {
		return err
	}
	return nil
}
