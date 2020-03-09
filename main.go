package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

	"github.com/gocolly/colly"
)

var auth smtp.Auth

type request struct {
	from    string
	to      []string
	subject string
	body    string
}

type htmlData struct {
	Link string
	Img  string
}

type templateData struct {
	Data []htmlData
}

func main() {
	m := templateData{}

	c := colly.NewCollector(
		colly.AllowedDomains("www.amctheatres.com"),
	)

	c.OnHTML("div[class=slick-list]", func(e *colly.HTMLElement) {
		links := e.ChildAttrs("a", "href")
		images := e.ChildAttrs("img", "src")
		uniqueImages := unique(images)
		uniqueLinks := unique(links)
		showtimesRemoved := []string{}
		for _, link := range uniqueLinks {
			if !strings.Contains(link, "showtimes") {
				link = fmt.Sprintf("https://www.amctheatres.com%s", link)
				showtimesRemoved = append(showtimesRemoved, link)
			}
		}

		out := []htmlData{}

		for i := range uniqueImages {
			t := htmlData{
				Link: showtimesRemoved[i],
				Img:  uniqueImages[i],
			}
			out = append(out, t)
		}

		m.Data = out
	})

	c.Visit("https://www.amctheatres.com/movie-theatres/phoenix/amc-ahwatukee-24")

	auth = smtp.PlainAuth("", "etokatlian@gmail.com", "", "smtp.gmail.com")
	r := newRequest([]string{"etokatlian@gmail.com", "tsmith93036@gmail.com"}, "This weeks movie briefing", "Hello, World!")
	err := r.ParseTemplate("template.html", m)
	if err := r.ParseTemplate("template.html", m); err == nil {
		ok, _ := r.SendEmail()
		fmt.Println(ok)
	}
	fmt.Println(err)
}

func newRequest(to []string, subject, body string) *request {
	return &request{
		to:      to,
		subject: subject,
		body:    body,
	}
}

func (r *request) SendEmail() (bool, error) {
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + r.subject + "!\n"
	msg := []byte(subject + mime + "\n" + r.body)
	addr := "smtp.gmail.com:587"

	if err := smtp.SendMail(addr, auth, "etokatlian@gmail.com", r.to, msg); err != nil {
		return false, err
	}
	return true, nil
}

func (r *request) ParseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	r.body = buf.String()
	return nil
}

func unique(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
