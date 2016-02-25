# The official Golang package for Kraken.io API

## Installation

Download and install the package

````
go get github.com/kraken-io/kraken-go
````

Add Kraken.io package to your project

````
import "github.com/kraken-io/kraken-go"
````

## Usage - Image URL

package main

import (
    "log"
    "github.com/kraken-io/kraken-go"
)

func main() {
    kr, err := kraken.New("your-api-key", "your-api-secret")

    if err != nil {
        log.Fatal(err)
    }

    params := map[string]interface {} {
        "wait": true,
        "url":  "http://img.rezeptebuch.com/thumbnail/Donauwelle-20-Cupcakes.jpg"
    }

    data, err := kr.URL(params)

    if err != nil {
        log.Fatal(err)
    }

    if data["success"] != true {
        log.Println("Failed, error message ", data["message"])
    } else {
        log.Println("Success, Optimized image URL: ", data["kraked_url"])
    }
}
        
## Usage - Image Upload

package main

import (
    "log"
    "github.com/kraken-io/kraken-go"
)

func main() {
    kr, err := kraken.New("your-api-key", "your-api-secret")

    if err != nil {
        log.Fatal(err)
    }

    params := map[string]interface {} {
        "wait": true
    }

    imgPath := "./path/to/file/on/disk.jpg"

    data, err := kr.Upload(params, imgPath)

    if err != nil {
        log.Fatal("err ", err)
    }
    
    if data["success"] != true {
        log.Println("Failed, error message ", data["message"])
    } else {
        log.Println("Success, Optimized image URL: ", data["kraked_url"])
    }
}
