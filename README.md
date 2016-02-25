# kraken-go
### Installation  
`$go get github.com/kraken-io/kraken-go`  
Then add to your project:  
`import "github.com/kraken-io/kraken-go"`  
### Usage - Image URL  
        kr, err := kraken.New("api_key", "api_secret")
        if err != nil {
                log.Fatal(err)
        }
        params := map[string]interface{}{
                "wait": true,
                "url":  "http://image-url.com/file.jpg",
                "resize": map[string]interface{}{
                        "width":    100,
                        "height":   75,
                        "strategy": "crop",
                },
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
### Usage - Image upload  
        kr, err := kraken.New("api_key", "api_secret")
        if err != nil {
                log.Fatal(err)
        }
        params := map[string]interface{}{
                "wait": true,
                "resize": map[string]interface{}{
                        "width":    100,
                        "height":   75,
                        "strategy": "crop",
                },
        }
        
        imgPath := "/path_to_file.jpg"
        data, err := kr.Upload(params, imgPath)
        if err != nil {
                t.Fatal("err ", err)
        }
        if data["success"] != true {
                log.Println("Failed, error message ", data["message"])
        } else {
				log.Println("Success, Optimized image URL: ", data["kraked_url"])
        }  
