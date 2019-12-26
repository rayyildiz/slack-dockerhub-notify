build:
	gcloud app deploy app.yaml --project rayyildiz-playground


gcloud:
	docker build -t eu.gcr.io/$PROJECT_ID/slack-nofify .