// Based on https://www.kluenter.de/garmin-ephemeris-files-and-linux/ and
// EPO_Downloader.rb in https://github.com/scrapper/postrunner (GPLv2)

// = EPO_Downloader.rb -- PostRunner - Manage the data from your Garmin sport devices.
//
// Copyright (c) 2015 by Chris Schlaeger <cs@taskjuggler.org>
//
// armstrong.go
//
// Copyright (c) 2016 by Steven Maude
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of version 2 of the GNU General Public License as
// published by the Free Software Foundation.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"flag"
)

// retrieveData makes a HTTP request to get Garmin EPO data and returns the body as []byte if successful.
func retrieveDataEPO() ([]byte, error) {
	url := "https://omt.garmin.com/Rce/ProtobufApi/EphemerisService/GetEphemerisData"
	// Data from https://www.kluenter.de/garmin-ephemeris-files-and-linux/
	data := []byte("\n-\n\aexpress\u0012\u0005de_DE\u001A\aWindows\"" +
		"\u0012601 Service Pack 1\u0012\n\b\x8C\xB4\x93\xB8" +
		"\u000E\u0012\u0000\u0018\u0000\u0018\u001C\"\u0000")

	c := &http.Client{
		Timeout: 20 * time.Second,
	}
	resp, err := c.Post(url, "application/octet-stream", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// checkDataLengthEPO checks the EPO data length; if not as expected, returns an error.
func checkDataLengthEPO(data []byte) error {
	dataLength := len(data)
	// Each EPO data set is 2307 bytes long, with the first three bytes to be removed.
	if dataLength != 28*2307 {
		return fmt.Errorf("EPO data has unexpected length: %v", dataLength)
	}
	return nil
}

// cleanEPO removes the first three bytes from each block of 2307 bytes in EPO data,
// and returns a cleaned []byte.
func cleanEPO(rawEPOData []byte) []byte {
	var outData []byte
	for i := 0; i <= 27; i++ {
		offset := i * 2307
		outData = append(outData, rawEPOData[offset+3:offset+2307]...)
	}
	return outData
}

// Retrieves EPO data, checks it, cleans it and writes it to disk.
func downloadFileEPO() {
	fmt.Println("Retrieving EPO data from Garmin's servers...")
	rawEPOData, err := retrieveDataEPO()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Processing EPO.BIN...")
	err = checkDataLengthEPO(rawEPOData)
	if err != nil {
		log.Fatal(err)
	}

	outData := cleanEPO(rawEPOData)

	err = ioutil.WriteFile("EPO.BIN", outData, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done! EPO.BIN saved.")
}


// retrieveData makes a HTTP request to get Garmin EPO data and returns the body as []byte if successful.
func retrieveDataCPE() ([]byte, error) {
	url := "https://api.gcs.garmin.com/ephemeris/cpe/sony?coverage=WEEKS_1"

	c := &http.Client{
		Timeout: 20 * time.Second,
	}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// checkDataLengthCPE errors out of CPE data is empty
// - CPE internal data format is not yet known
func checkDataLengthCPE(data []byte) error {
	dataLength := len(data)
	if dataLength == 0 {
		return fmt.Errorf("CPE data has unexpected length of zero")
	}
	return nil
}

// Retrieves CPE data, checks it and writes it to disk.
func downloadFileCPE() {
	fmt.Println("Retrieving CPE data from Garmin's servers...\n")
	rawCPEData, err := retrieveDataCPE()
	if err != nil {
		log.Fatal(err)
	}

	err = checkDataLengthCPE(rawCPEData)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("CPE.BIN", rawCPEData, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done! CPE.BIN saved.\n")
}


func main() {
	var modeCPE bool
	flag.BoolVar(&modeCPE, "cpe", false, "Download CPE format ephemeris data (instead of EPO)")
	flag.Parse()

	if modeCPE == true {
		downloadFileCPE()
	} else {
		downloadFileEPO()
	}
}
