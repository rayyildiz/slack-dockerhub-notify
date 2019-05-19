package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

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
	_, err = client.Post(webhookURL, "application/json", b)

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
		<title>Slack DockerHub Notification</title>
		<meta name="keywords" content="slack,golang,notification,docker,docker hub">
		<link rel="stylesheet" href="//unpkg.com/purecss@1.0.0/build/pure-min.css" integrity="sha384-nn4HPE8lTHyVtfCBi5yW9d20FjT8BJwUXyWZT9InLYax14RDjBj46LmSztkmNP9w" crossorigin="anonymous">
		<script src="//cdnjs.cloudflare.com/ajax/libs/showdown/1.8.2/showdown.min.js" type="text/javascript"></script>
		<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/3.2.1/jquery.min.js" type="text/javascript"></script>
	</head>
	<body>
		<div style="margin:1em">

<form class="pure-form pure-form-aligned">
		<h1>Create a webhook URL.</h1>
		<h4>Enter your slack webhook url and click convert button</h4>
		<div class="pure-control-group">
			<input class="pure-input-2-3" id="input_url" type="text" placeholder="Enter Slack Web Income URL. example https://hooks.slack.com/services/T123456789/000000000000000/777777777777777777777">
			<a href="#" onclick="javascript:convertURL()" class="pure-button pure-button-primary pure-input-1-3">Convert</a>
		</div>
		<p id="error_text"></p>
		<div id="div_result" class="pure-control-group hidden">
			<h3>Create a webhook on dockerhub and use this url: </h3>
			<input class="pure-input-1" readonly id="result_url" type="text" placeholder="Result">
		</div>
</form>
			<br /><hr /> <br />

			<div id="html" style="visibility:visible"></div>
		</center>
			<div id="markdown" style="visibility:hidden">
{{.}}
			</div>

			<script type="text/javascript">

				function convertURL() {
						$('#div_result').attr('class','pure-control-group hidden');
						$('#error_text').html(' ');

						var text = $('#input_url').val();
						ind = text.indexOf("https://hooks.slack.com/services/")
						if (ind != 0 ) {
							$('#error_text').html('URL should like <b>https://hooks.slack.com/services/T123456789/000000000000000/777777777777777777777</b>');
						} else {
							rest = text.substring(ind+23);

							$('#result_url').val(window.location.protocol + '//' + window.location.hostname + rest);
							$('#div_result').attr('class','pure-control-group');
						}
				}
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

	if err := tmpl.Execute(w, fmt.Sprintf("%s", data)); err != nil {
		log.Errorf("could not execute template, %v", err)
	}

}

func main() {
	log.SetOutput(os.Stdout)
	// log.SetFormatter(&log.JSONFormatter{})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("Starting appengine")
	http.HandleFunc("/", handler)
	log.Println("Started appengine")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("could not start server :%s, %v", port, err)
	}

}
