package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"time"
)

type Data struct {
	Date        string
	Time        string
	Description string
	Department  string
}

type State struct {
	InFieldset bool
	InTable    bool
	InTbody    bool
	InTd       bool
	TagName    string
	tt         html.TokenType
}

func main() {
	lastState := 0
	for {
		data := getPackageState()
		dataSize := len(data)

		if lastState < dataSize {
			buff := bytes.NewBufferString("")
			_, _ = fmt.Fprintf(buff, "%v", data)
			_ = sendMail(buff.String())
			lastState = dataSize
		}

		time.Sleep(time.Minute * 10)
	}
}

func getPackageState() []Data {
	v := url.Values{}
	v.Add("q", "0000211972267U")
	v.Add("typ", "1")

	res, err := http.PostForm("https://tracktrace.dpd.com.pl/findPackage", v)

	if err != nil {
		log.Printf("http request error: %s", err)
	}

	tokenizer := html.NewTokenizer(res.Body)
	state := State{}
	var data []string

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		tagName, _ := tokenizer.TagName()
		state.TagName = string(tagName)

		if tt == html.EndTagToken && state.TagName == "fieldset" {
			state.InFieldset = false
		}

		if state.InFieldset {
			data = fieldset(tt, tokenizer, &state, data)
		}

		if tt == html.StartTagToken && state.TagName == "fieldset" {
			state.InFieldset = true
		}
	}

	var dane []Data
	var history Data

	for i, datum := range data {
		switch i % 4 {
		case 0:
			history = Data{}
			history.Date = datum
			break
		case 1:
			history.Time = datum
			break
		case 2:
			history.Description = datum
			break
		case 3:
			history.Department = datum
			dane = append(dane, history)
		}
	}

	return dane
}

func fieldset(tt html.TokenType, tokenizer *html.Tokenizer, state *State, data []string) []string {
	if tt == html.EndTagToken && state.TagName == "table" {
		state.InTable = false
	}

	if state.InTable {
		if tt == html.EndTagToken && state.TagName == "tbody" {
			state.InTbody = false
		}

		if state.InTbody {
			if tt == html.EndTagToken && state.TagName == "td" {
				state.InTd = false
			}

			if state.InTd {
				data = append(data, tokenizer.Token().Data)
			}

			if tt == html.StartTagToken && state.TagName == "td" {
				state.InTd = true
			}
		}

		if tt == html.StartTagToken && state.TagName == "tbody" {
			state.InTbody = true
		}
	}

	if tt == html.StartTagToken && state.TagName == "table" {
		state.InTable = true
	}

	return data
}

func sendMail(content string) error {
	from := os.Getenv("EMAIL_FROM")
	pass := os.Getenv("EMAIL_PASS")
	to := os.Getenv("EMAIL_TO")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Paczka DPD\n" +
		"MIME-Version: 1.0\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\n" +
		"Content-Transfer-Encoding: base64\n\n" +
		base64.StdEncoding.EncodeToString([]byte(content))

	return smtp.SendMail(smtpHost+":"+smtpPort,
		smtp.PlainAuth("", from, pass, smtpHost),
		from,
		[]string{to},
		[]byte(msg))
}
