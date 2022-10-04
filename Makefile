build:
	gcloud app deploy app.yaml --project $PROJECT_ID


gcloud:
	docker build -t eu.gcr.io/$PROJECT_ID/slack-nofify .