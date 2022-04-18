package main

import "errors"

func RequiredArg(argName string) error {
	return errors.New("arg: " + argName + " is required")
}

func InvalidArg(argName string) error {
	return errors.New("arg: " + argName + " is invalid")
}

var sdkNotInitialized = errors.New("SDK is not initialised")
