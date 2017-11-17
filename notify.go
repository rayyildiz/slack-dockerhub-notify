package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"log"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

type Attachment struct {
	Color string `json:"color"`
	Text  string `json:"text"`
}

type Payload struct {
	Attachments []Attachment `json:"attachments,omitempty"`
}

// DockerHub details.
type DockerHub struct {
	PushedData struct {
		Tag string `json:"tag,omitempty"`
	} `json:"push_data"`
	Repository struct {
		Status   string `json:"status,omitempty"`
		RepoName string `json:"repo_name,omitempty"`
		RepoURL  string `json:"repo_url,omitempty"`
	} `json:"repository"`
}

func send(ctx context.Context, webhookURL string, payload Payload) error {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(payload)
	if err != nil {
		return fmt.Errorf("encode problem %v", err)
	}

	client := urlfetch.Client(ctx)
	res, err := client.Post(webhookURL, "application/json", b)
	if res.StatusCode >= 400 {
		return fmt.Errorf("Error sending msg. Status: %v", res.Status)
	}

	return err
}

func handler(w http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {
		homepageHandler(w, req)
		return
	}

	if req.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Not supported web method %s", req.Method), http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var formData DockerHub
	err := decoder.Decode(&formData)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	payload := Payload{}
	var attachment Attachment
	attachment.Text = fmt.Sprintf("<%s|%s:%s> build status: %s", formData.Repository.RepoURL, formData.Repository.RepoName, formData.PushedData.Tag, formData.Repository.Status)

	if formData.Repository.Status == "Active" {
		attachment.Color = "#36a64f"
	} else {
		attachment.Color = "#FF3118"
	}

	payload.Attachments = append(payload.Attachments, attachment)

	url := fmt.Sprintf("https://hooks.slack.com/%s", req.URL.Path)
	ctx := appengine.NewContext(req)
	err = send(ctx, url, payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "message was sent to slack")
}

var tmpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="en">
	<head>
    <meta charset="utf-8">
		<meta name="viewport" content="width=device-width">
		<title>Slack Notification.</title>
		<link rel="stylesheet" href="//unpkg.com/purecss@1.0.0/build/pure-min.css" integrity="sha384-nn4HPE8lTHyVtfCBi5yW9d20FjT8BJwUXyWZT9InLYax14RDjBj46LmSztkmNP9w" crossorigin="anonymous">
		<script src="//cdnjs.cloudflare.com/ajax/libs/showdown/1.8.2/showdown.min.js" type="text/javascript"></script>
		<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/3.2.1/jquery.min.js" type="text/javascript"></script>

			<!--[if lte IE 8]>
          <link rel="stylesheet" href="//unpkg.com/purecss@1.0.0/build/grids-responsive-old-ie-min.css">
      <![endif]-->
      <!--[if gt IE 8]><!-->
          <link rel="stylesheet" href="//unpkg.com/purecss@1.0.0/build/grids-responsive-min.css">
      <!--<![endif]-->

		<script async src="//www.googletagmanager.com/gtag/js?id=UA-49404964-3"></script>
		<script>
		  window.dataLayer = window.dataLayer || [];
		  function gtag(){dataLayer.push(arguments);}
		  gtag('js', new Date());

		  gtag('config', 'UA-49404964-3');
		</script>

	</head>
	<body>
		<div  class="custom-wrapper" style="margin:1em">
			<div id="html" style="visibility:visible"></div>
		</center>
			<div id="markdown" style="visibility:hidden">{{.}}</div>

			<script type="text/javascript">
				$( document ).ready(function() {
					 var text = $('#markdown').html();
					 converter = new showdown.Converter();
					 html = converter.makeHtml(text);
					 $('#html').html(html);
				});
			</script>
	</body>
</html>
`))

func homepageHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadFile("README.md")
	if err != nil {
		http.Error(w, fmt.Sprintf("Not supported web method %s", req.Method), http.StatusInternalServerError)

		return
	}

	tmpl.Execute(w, fmt.Sprintf("%s", data))
}

func init() {
	log.Println("Starting appengine")
	http.HandleFunc("/", handler)
	log.Println("Started appengine")
}
