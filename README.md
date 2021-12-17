# Zwift Power results automation


Grab latest results from our team members from ZwiftPower

Runs in Google Cloud Run as an http service. Trigger a new run with:

```bash
curl -H \                       
"Authorization: Bearer $(gcloud auth print-identity-token)" \
https://<service URL>/trigger
```

Environment variables on the Google Cloud Run service:

* SPREADSHEET_ID: Google sheets ID
* SPREADSHEET_SHEET: Name of the sheet
* LIMIT: for testing, limit the number of riders we get data for

If you don't set SPREADSHEET_ID, you get the results written to a results.csv file in the Google Cloud storage bucket. 
