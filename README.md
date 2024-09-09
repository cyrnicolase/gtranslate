# gtranslate ![build](https://travis-ci.com/bregydoc/gtranslate.svg?branch=master)

Google Translate API for unlimited and free translations 📢.
This project was inspired by [google-translate-api](https://github.com/matheuss/google-translate-api) and [google-translate-token](https://github.com/matheuss/google-translate-token).

# Install

    go get github.com/cyrnicolase/gtranslate

# Example

```go
package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/cyrnicolase/gtranslate"
	"golang.org/x/text/language"
)

type Content struct {
	Text string
	From language.Tag
	To   language.Tag
}

var (
	cc = []Content{
		{
			Text: "Hello World",
			From: language.English,
			To:   language.Chinese,
		},
		{
			Text: "اطلع على الكلمات كاملةً",
			From: language.Arabic,
			To:   language.Chinese,
		},
		{
			Text: "床前明月光，疑是地上霜",
			From: language.Chinese,
			To:   language.English,
		},
		{
			Text: "危楼高百尺，手可摘星辰\n不敢高声语，恐惊天上人",
			From: language.Chinese,
			To:   language.Arabic,
		},
	}
)

func main() {
	tr := gtranslate.NewTranslate()
	wg := new(sync.WaitGroup)
	for _, c := range cc {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()

			result, err := tr.Run(context.TODO(), c.Text, c.From, c.To)
			if err != nil {
				log.Fatal(err)
			}

			duration := time.Now().Sub(start).Milliseconds()
			log.Printf("Original Text: %s \nTranslated Text: %s \nTranslated Tongue: %s \ntimeDuration: %d millis\n\n\n", c.Text, result.ResponseText, result.ResponseTongue, duration)
		}()
	}
	wg.Wait()
}

```
