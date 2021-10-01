package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/google/go-github/v39/github"
	"github.com/kkyr/fig"
	"github.com/otiai10/copy"
	"golang.org/x/image/draw"
	"golang.org/x/oauth2"
)

type config struct {
	OAuth      string   `fig:"oauth" validate:"required"`
	Org        string   `fig:"org" default:"shopware"`
	Repo       string   `fig:"repo" default:"platform"`
	Excluded   []string `fig:"excluded" default:"[]"`
	AvatarSize int      `fig:"avatarSize" default:"100"`
}

type swagContributor struct {
	Name       string
	ProfileURL string
	AvatarURL  string
}

type PageData struct {
	Org          string
	Repo         string
	Contributors []swagContributor
}

var excludes map[string]struct{}
var cfg config

func main() {
	err := fig.Load(&cfg,
		fig.File("config.yaml"),
		fig.Dirs("."),
	)
	if err != nil {
		displayHelp()
		fmt.Println(err.Error())
		os.Exit(1)
	}

	excludes = make(map[string]struct{})
	for _, login := range cfg.Excluded {
		excludes[login] = struct{}{}
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.OAuth},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	opt := &github.ListContributorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allContribs []*github.Contributor
	for {
		contribs, resp, err := client.Repositories.ListContributors(context.Background(), cfg.Org, cfg.Repo, opt)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		allContribs = append(allContribs, contribs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	var scs []swagContributor

	for _, contrib := range allContribs {
		if _, ok := excludes[contrib.GetLogin()]; ok {
			continue
		}
		scs = append(scs, swagContributor{
			Name:       contrib.GetLogin(),
			ProfileURL: contrib.GetHTMLURL(),
			AvatarURL:  contrib.GetAvatarURL(),
		})
	}
	fmt.Println("Number of contributors: ", len(allContribs))
	fmt.Println("Number of contributors not excluded: ", len(scs))
	fmt.Println("Downloading avatars")
	for _, sc := range scs {
		err := downloadFile(sc.AvatarURL, sc.Name)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	fmt.Println("Building HTML output")

	tmpl := template.Must(template.ParseFiles("template.gohtml"))
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, PageData{
		Org:          cfg.Org,
		Repo:         cfg.Repo,
		Contributors: scs,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	_ = os.RemoveAll("output")
	err = os.MkdirAll("output", 0755)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = ioutil.WriteFile("output/contributors.html", buf.Bytes(), 0666)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = copy.Copy("imgCache", "output/avatars")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Done\nOutput is in folder output")
}

func downloadFile(url, name string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Received non 200 response code: %d", response.StatusCode))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not read avatar for %s from %s", name, url))
	}
	cType := http.DetectContentType(body)
	var srcImg image.Image
	switch cType {
	case "image/jpg":
		srcImg, err = jpeg.Decode(bytes.NewReader(body))
		if err != nil {
			return err
		}
	case "image/jpeg":
		srcImg, err = jpeg.Decode(bytes.NewReader(body))
		if err != nil {
			return err
		}
	case "image/png":
		srcImg, err = png.Decode(bytes.NewReader(body))
		if err != nil {
			return err
		}
	}
	rect := image.Rect(0, 0, cfg.AvatarSize, cfg.AvatarSize)
	img := image.NewRGBA(rect)
	draw.BiLinear.Scale(img, rect, srcImg, srcImg.Bounds(), draw.Over, nil)
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return err
	}
	return ioutil.WriteFile(fmt.Sprintf("imgCache/%s.jpg", name), buf.Bytes(), 0666)
}

func displayHelp() {
	fmt.Print(`
Could not read your github oauth token.
Did you set it in config.yaml?

If you need to generate a token follow these instructions:
https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token

After that, rename config.yaml.dist to config.yaml and set the token in the file.

`)
}
